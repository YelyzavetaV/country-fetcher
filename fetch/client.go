package fetch

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/YelyzavetaV/country-fetcher/models"
	"github.com/YelyzavetaV/country-fetcher/process"
)

type Client interface {
	FetchCountries(query Query, n int) ([]models.Country, error)
	ProcessRegion(name string, n int) (*models.Region, error)
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

func (c *clientImpl) ProcessRegion(name string, n int) (*models.Region, error) {
	query := RegionQuery{name}

	countries, err := c.FetchCountries(query, n)
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