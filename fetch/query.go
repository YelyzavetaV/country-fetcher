package fetch

import (
	"fmt"

	"github.com/YelyzavetaV/country-fetcher/config"
)

type Query interface {
	buildURL() string
}

type NameQuery struct {
	Name     string
	FullText bool
}

type CodeQuery struct {
	Code string
}

type RegionQuery struct {
	Region string
}

func (q NameQuery) buildURL() string {
	return fmt.Sprintf(
		"%s/name/%s?fullText=%t", config.BaseURL, q.Name, q.FullText)
}

func (q RegionQuery) buildURL() string {
	return fmt.Sprintf("%s/region/%s", config.BaseURL, q.Region)
}

func (q CodeQuery) buildURL() string {
	return fmt.Sprintf("%s/alpha/%s", config.BaseURL, q.Code)
}