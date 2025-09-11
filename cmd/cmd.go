package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/YelyzavetaV/country-fetcher/fetch"
)

var client fetch.Client = fetch.NewClient()

// Command-line arguments
var (
	name     string
	region   string
	code     string
	fullText bool
	all      bool
	n        int
)

var RootCmd = &cobra.Command{
	Use:   "country-fetcher",
	Short: "Fetch country information from REST Countries",
	Long:  `
A CLI tool to fetch and process country and region data using RESTful Country API.
	`,
}

var fetchCountriesCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Get info about a country or countries",
	Long:  `
Get info about one or multiple countries by country name, region name, or country code.

Arguments:
--all: fetch all matching countries - typically, in a given region.
	   Note that --all has precendence over --n, i.e., providing
	   "--all true --n <value>" will lead to the value of n being
	   ignored.
	`,
	Run: fetchCountries,
}

var processRegionCmd = &cobra.Command{
	Use:   "region",
	Short: "Get stats of n or all countries in a region",
	Long:  "",
	Run:   processRegion,
}

func init() {
	fetchCountriesCmd.Flags().StringVar(&name, "name", "", "Country name")
	fetchCountriesCmd.Flags().StringVar(
		&region, "region", "", "Geographical region")
	fetchCountriesCmd.Flags().StringVar(&code, "code", "", "Country code")
	fetchCountriesCmd.Flags().BoolVar(
		&fullText, "fulltext", false, "Use full text name search")
	fetchCountriesCmd.Flags().BoolVar(
		&all, "all", false,
			"Fetch all matching countries (typically, in a region)")
	fetchCountriesCmd.Flags().IntVar(
		&n, "n", 1, "Maximum number of countries to fetch")

	RootCmd.AddCommand(fetchCountriesCmd)

	processRegionCmd.Flags().StringVar(&name, "name", "", "Region name")
	processRegionCmd.Flags().BoolVar(
		&all, "all", false, "Fetch all countries in the region")
	processRegionCmd.Flags().IntVar(
		&n, "n", 10, "Maximum number of countries to fetch")

	RootCmd.AddCommand(processRegionCmd)
}

func fetchCountries(cmd *cobra.Command, args []string) {
	var query fetch.Query
	if name != "" {
		query = fetch.NameQuery{
			Name:     name,
			FullText:fullText,
		}
	} else if region != "" {
		query = fetch.RegionQuery{
			Region: region,
		}
	} else if code != "" {
		query = fetch.CodeQuery{
			Code: code,
		}
	} else {
		return
	}

	if all { n = -1 }
	countries, err := client.FetchCountries(query, n)
	if err != nil {
		fmt.Printf("Failed to fetch data: %v", err)
		return
	}

	fmt.Println("Fetched data:")
	for _, country := range countries {
		fmt.Printf("%+v\n", country)
	}
}

func processRegion(cmd *cobra.Command, args []string) {
	if all { n = -1 }

	region, err := client.ProcessRegion(name, n)
	if err != nil {
		fmt.Printf("Failed to get region data: %v", err)
		return
	}

	fmt.Printf("Region: %s\n", region.Name)
	fmt.Printf("Total population: %d\n", region.TotalPopulation)
	fmt.Printf("Average population: %f\n", region.AvgPopulation)

	countryNames := make([]string, len(region.Countries))
	for i, country := range region.Countries {
		countryNames[i] = country.Name
	}
	fmt.Printf("Countries: %v\n", strings.Join(countryNames, ", "))
}