package fetch

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/YelyzavetaV/country-fetcher/models"
)

type Client interface {
	Fetch(query Query, n int) ([]models.Country, error)
}

type clientImpl struct{}

func NewClient() Client {
	return &clientImpl{}
}

func (c *clientImpl) Fetch(q Query, n int) ([]models.Country, error) {
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

	if len(countries) > n {
		countries = countries[:n]
	}

	return countries, nil
}