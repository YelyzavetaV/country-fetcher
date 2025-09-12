package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/YelyzavetaV/country-fetcher/config"
	"github.com/YelyzavetaV/country-fetcher/client"
	"github.com/YelyzavetaV/country-fetcher/output"
)

var (
	cfg *config.Config
	c client.Client
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

var fetchRegionsCmd = &cobra.Command{
	Use:   "fetch-regions",
	Short: "Get stats of n or all countries in the region",
	Long:  `
Compute statistics for n or all countries in the region.
	`,
	Run:   fetchRegions,
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

	fetchRegionsCmd.Flags().StringSliceVar(
		&regions, "regions", []string{},
			"Comma-separated list of region names (e.g., Europe,Asia,Africa)")
	fetchRegionsCmd.MarkFlagRequired("regions")

	fetchRegionsCmd.Flags().BoolVar(
		&all, "all", false,
			"Fetch all countries in the region. Note that --all has " +
			"precendence over --n, i.e., providing '--all true --n <value>' " +
			"will lead to the value of n being ignored.")
	fetchRegionsCmd.Flags().IntVar(
		&n, "n", 10,
			"Maximum number of countries to fetch. For all=false, " +
			"a non-positive n is interpreted as all=true.")
	fetchRegionsCmd.Flags().StringVar(
		&filename, "file", "",
			"Name of a file the output is to be written to. If not " +
			"provided, JSON string is outputted to console.")

	RootCmd.AddCommand(fetchRegionsCmd)
}

func setup(cmd *cobra.Command, args []string) {
	cfg = config.NewConfig()
	c = client.NewClient()
}

func fetchCountries(cmd *cobra.Command, args []string) {
	// Assemble queries.
	var queries []client.Query
	if len(names) != 0 {
		for _, name := range names {
			queries = append(
				queries,
				client.NameQuery{
					Name:     name,
					FullText: fullText,
				},
			)
		}
	} else if len(codes) != 0 {
		for _, code := range codes {
			queries = append(queries, client.CodeQuery{Code: code})
		}
	} else {
		return
	}

	if all { n = -1 }
	ch := client.Fetch(queries, c.FetchCountries, n)

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

func fetchRegions(cmd *cobra.Command, args []string) {
	queries := make([]client.Query, len(regions))

	if all { n = -1 }
	ch := client.Fetch(queries, c.FetchRegion, n)

	for range queries {
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