package client

import (
	"fmt"
	"encoding/json"
	"io"
	"net/http"

	"github.com/YelyzavetaV/country-fetcher/models"
	"github.com/YelyzavetaV/country-fetcher/process"
)

// Provides methods for fetching country and region data via RESTful
// Country API
type Client interface {
	// Fetch n or all countries matching the query
	// (by country name, by region name, or by country code).
	FetchCountries(q Query, n int) ([]models.Country, error)

	// Compute region statistics for n or all countries.
	FetchRegion(q Query, n int) (*models.Region, error)
}

type clientImpl struct{}

func NewClient() Client {
	return &clientImpl{}
}

func (c *clientImpl) FetchCountries(q Query, n int) ([]models.Country, error) {
	url := q.buildURL()

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
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

	// Setting n to -1 lets fetching all matching countries, for instance,
	// all countries in a region
	if n > 0 && len(countries) > n {
		countries = countries[:n]
	}

	return countries, nil
}

func (c *clientImpl) FetchRegion(q Query, n int) (*models.Region, error) {
	// Validate that the input query is RegionQuery
	var name string
	if rq, ok := q.(RegionQuery); ok {
		name = rq.Region
	} else {
		return nil, fmt.Errorf(
			"FetchRegion requires RegionQuery; got %T", q)
	}

	countries, err := c.FetchCountries(q, n)
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
func Fetch[T any](
	queries []Query,
	fetcher func(Query, int) (T, error),
	n int,
) chan T {
	ch := make(chan T, len(queries))

	for _, q := range queries {
		go func(query Query) {
			res, err := fetcher(query, n)
			if err != nil {
				fmt.Printf("Failed to fetch data: %v\n", err)

				var zero T
				ch <- zero
				return
			}
			ch <- res
		}(q)
	}
	return ch
}