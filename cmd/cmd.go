package cmd

import (
	"fmt"

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
	n        int
)

var RootCmd = &cobra.Command{
	Use:   "country-fetcher",
	Short: "Fetch country information from REST Countries",
	Long:  `
A CLI tool to fetch and process country and region data using RESTful Country API.
	`,
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Get info about a country or countries",
	Long:  `
Get info about one or multiple countries by country name, region name, or country code.
	`,
	Run: fetchCountries,
}

func init() {
	fetchCmd.Flags().StringVar(&name, "name", "", "Country name")
	fetchCmd.Flags().StringVar(&region, "region", "", "Geographical region")
	fetchCmd.Flags().StringVar(&code, "code", "", "Country code")
	fetchCmd.Flags().BoolVar(
		&fullText, "fulltext", false, "Use full text name search")
	fetchCmd.Flags().IntVar(
		&n, "n", 1, "Maximum number of countries to fetch")

	RootCmd.AddCommand(fetchCmd)
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