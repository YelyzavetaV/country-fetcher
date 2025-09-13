package client

import (
	"fmt"
	"encoding/json"
	"io"
	"net/http"
	"context"
	"time"
	"sync"

	"github.com/YelyzavetaV/country-fetcher/models"
	"github.com/YelyzavetaV/country-fetcher/process"
)

// Abstracts the response of client fetch attempt
type FetchResponse struct {
	Value interface{}
	Err   error
}

// Provides methods for fetching country and region data via RESTful
// Country API
type Client interface {
	// Fetch n or all countries matching the query
	// (by country name, by region name, or by country code).
	fetchCountries(
		q Query, ncMax int, timeout time.Duration) ([]models.Country, error)

	// Compute region statistics for n or all countries.
	fetchRegion(
		q Query, ncMax int, timeout time.Duration) (*models.Region, error)

	// Fetch data for multiple countries or multiple regions concurrently.
	Fetch(queries []Query, ncMax int, timeout time.Duration) chan FetchResponse
}

type clientImpl struct{}

func NewClient() Client {
	return &clientImpl{}
}

func (c *clientImpl) fetchCountries(
	q Query, ncMax int, timeout time.Duration,
) ([]models.Country, error) {
	url := q.buildURL()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var countries []models.Country
	if err := json.Unmarshal(body, &countries); err != nil {
		// If fetching by code, we expect a single-valued response
		var country models.Country
		if e := json.Unmarshal(body, &country); e != nil {
			return nil, e
		}

		countries = append(countries, country)
	}

	if len(countries) == 0 {
		return nil, err
	}

	// Setting ncMax to -1 lets fetching all matching countries, for instance,
	// all countries in a region
	if ncMax > 0 && len(countries) > ncMax {
		countries = countries[:ncMax]
	}

	return countries, nil
}

func (c *clientImpl) fetchRegion(
	q Query, ncMax int, timeout time.Duration,
) (*models.Region, error) {
	// Validate that the input query is RegionQuery
	var name string
	if rq, ok := q.(RegionQuery); ok {
		name = rq.Region
	} else {
		return nil, fmt.Errorf(
			"fetchRegion requires RegionQuery; got %T", q)
	}

	countries, err := c.fetchCountries(q, ncMax, timeout)
	if err != nil {
		return nil, err
	}

	nc := len(countries)
	if nc == 0 {
		return nil, err
	}

	region := models.Region{
		Name:      name,
		Countries: countries,
	}

	p := make([]int, nc)
	for i, country := range region.Countries {
		p[i] = country.Population
	}

	region.TotalPopulation = process.Sum(p)
	region.AvgPopulation = float64(region.TotalPopulation) / float64(nc)

	return &region, nil
}

// Execute fetch queries each in their own goroutine.
func (c *clientImpl) Fetch(
	queries []Query, n int, timeout time.Duration,
) chan FetchResponse {
	var f func(Query, int) (interface{}, error)
	if _, ok := queries[0].(RegionQuery); ok {
		f = func(q Query, n int) (interface{}, error) {
			return c.fetchRegion(q, n, timeout)
		}
	} else {
		f = func(q Query, n int) (interface{}, error) {
			return c.fetchCountries(q, n, timeout)
		}
	}

	nq := len(queries)

	// Make buffered channel and wait group to eventually close ch
	ch := make(chan FetchResponse, nq)
	var wg sync.WaitGroup
	wg.Add(nq)

	for _, q := range queries {
		go func(query Query) {
			defer wg.Done()

			res, err := f(query, n)
			ch <- FetchResponse{Value: res, Err: err}
		}(q)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}