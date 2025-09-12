package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/YelyzavetaV/country-fetcher/config"
	"github.com/YelyzavetaV/country-fetcher/fetch"
	"github.com/YelyzavetaV/country-fetcher/output"
	"github.com/YelyzavetaV/country-fetcher/models"
)

var (
	cfg *config.Config
	client fetch.Client
)

// Command-line arguments
var (
	names    []string
	codes    []string
	regions  []string
	fullText bool
	all      bool
	n        int
	filename string
)

var RootCmd = &cobra.Command{
	Use:              "country-fetcher",
	Short:            "Fetch country information from REST Countries",
	Long:             `
A CLI tool to fetch and process country and region data using RESTful Country API.
	`,
	PersistentPreRun: setup,
}

var fetchCountriesCmd = &cobra.Command{
	Use:   "fetch-countries",
	Short: "Get info about a country or countries",
	Long:  `
Get info about one or multiple countries by country name or country code.
	`,
	Run:   fetchCountries,
}

var processRegionsCmd = &cobra.Command{
	Use:   "fetch-regions",
	Short: "Get stats of n or all countries in the region",
	Long:  `
Compute statistics for n or all countries in the region.
	`,
	Run:   processRegions,
}

func init() {
	fetchCountriesCmd.Flags().StringSliceVar(
		&names, "names", []string{},
			"Country names. Has precedence over --codes - providing " +
			"a non-empty --names leads to country codes being ignored.")
	fetchCountriesCmd.Flags().StringSliceVar(
		&codes, "codes", []string{}, "Country codes")
	fetchCountriesCmd.Flags().BoolVar(
		&fullText, "fulltext", false,
			"Use full text name search. Only relevant when quering country by name.")
	fetchCountriesCmd.Flags().IntVar(
		&n, "n", 1, "Maximum number of countries to fetch.")
	fetchCountriesCmd.Flags().StringVar(
		&filename, "file", "",
			"Name of a file the output is to be written to. If not " +
			"provided, JSON string is outputted to console.")

	RootCmd.AddCommand(fetchCountriesCmd)

	processRegionsCmd.Flags().StringSliceVar(
		&regions, "regions", []string{},
			"Comma-separated list of region names (e.g., Europe,Asia,Africa)")
	processRegionsCmd.MarkFlagRequired("regions")

	processRegionsCmd.Flags().BoolVar(
		&all, "all", false,
			"Fetch all countries in the region. Note that --all has " +
			"precendence over --n, i.e., providing '--all true --n <value>' " +
			"will lead to the value of n being ignored.")
	processRegionsCmd.Flags().IntVar(
		&n, "n", 10,
			"Maximum number of countries to fetch. For all=false, " +
			"a non-positive n is interpreted as all=true.")
	processRegionsCmd.Flags().StringVar(
		&filename, "file", "",
			"Name of a file the output is to be written to. If not " +
			"provided, JSON string is outputted to console.")

	RootCmd.AddCommand(processRegionsCmd)
}

func setup(cmd *cobra.Command, args []string) {
	cfg = config.NewConfig()
	client = fetch.NewClient()
}

func fetchCountries(cmd *cobra.Command, args []string) {
	var queries []fetch.Query

	if len(names) != 0 {
		for _, name := range names {
			queries = append(
				queries,
				fetch.NameQuery{
					Name:     name,
					FullText: fullText,
				},
			)
		}
	} else if len(codes) != 0 {
		for _, code := range codes {
			queries = append(queries, fetch.CodeQuery{Code: code})
		}
	} else {
		return
	}

	if all { n = -1 }

	ch := make(chan []models.Country, len(queries))

	for _, q := range queries {
		go func(query fetch.Query){
			countries, err := client.FetchCountries(query, n)
			if err != nil {
				fmt.Printf("Failed to fetch data: %v", err)
				ch <- nil
				return
			}
			ch <- countries
		}(q)
	}

	for range queries {
		countries := <-ch
		if countries == nil {
			fmt.Println("Countries fetch failed. Skipping...")
			continue
		}

		if err := output.ToJSON(
			countries,
			filename,
			cfg.JSONPrefix,
			cfg.JSONIndent,
			cfg.JSONFilePermission,
		); err != nil {
			fmt.Printf("Failed to output data: %v", err)
		}
	}
}

func processRegions(cmd *cobra.Command, args []string) {
	if all { n = -1 }

	ch := make(chan *models.Region, len(regions))

	for _, r := range regions {
		go func(name string) {
			region, err := client.ProcessRegion(name, n)
			if err != nil {
				fmt.Printf(
					"Failed to get data for region %s: %v", name, err)
				ch <- nil
				return
			}
			ch <- region
		}(r)
	}

	for range regions {
		region := <-ch
		if region == nil {
			fmt.Println("Region fetch failed. Skipping...")
			continue
		}

		if err := output.ToJSON(
			region,
			filename,
			cfg.JSONPrefix,
			cfg.JSONIndent,
			cfg.JSONFilePermission,
		); err != nil {
			fmt.Printf("Failed to output data: %v", err)
		}
	}
}