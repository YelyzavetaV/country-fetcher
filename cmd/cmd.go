package cmd

import (
	"os"
	"fmt"
	"time"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/YelyzavetaV/country-fetcher/logging"
	"github.com/YelyzavetaV/country-fetcher/config"
	"github.com/YelyzavetaV/country-fetcher/client"
	"github.com/YelyzavetaV/country-fetcher/output"
)

var (
	cfg *config.Config
	c client.Client
// Configs to be validated at setup
	timeout time.Duration
)

// Command-line arguments
var (
	names    []string
	codes    []string
	regions  []string
	fullText bool
	all      bool
	ncMax    int
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
		&ncMax, "n", 1,
			"Maximum number of countries to fetch " +
			"(relevant for fuzzy search by name).")
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
		&ncMax, "n", 10,
			"Maximum number of countries per region. For all=false, " +
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

	// Parse and validate configs
	var err error

	timeout, err = time.ParseDuration(cfg.HTTPTimeout)
	if err != nil {
		panic(fmt.Errorf(
			"Invalid HTTP timeout value '%s' in config. Please use " +
			"a valid duration string (e.g., '10s', '500ms').",
			cfg.HTTPTimeout))
	}

	var logLevel slog.Level
	if logLevel, err = logging.ParseLogLevel(cfg.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	logging.InitLogger(slog.HandlerOptions{
		Level: logLevel,
	})

	logging.Log.Info(
		"Setup complete",
		"HTTPTimeout",
		timeout,
		"LogLevel",
		logLevel,
	)
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

	if all { ncMax = -1 }
	ch := c.Fetch(queries, ncMax, timeout)

	for res := range ch {
		if res.Err != nil {
			fmt.Printf("Fetch failed: %v\n", res.Err)
			continue
		}

		if err := output.ToJSON(
			res.Value,
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
	for i, region := range regions {
		queries[i] = client.RegionQuery{region}
	}

	if all { ncMax = -1 }
	ch := c.Fetch(queries, ncMax, timeout)

	for res := range ch {
		if res.Err != nil {
			fmt.Printf("Fetch failed: %v\n", res.Err)
			continue
		}

		if err := output.ToJSON(
			res.Value,
			filename,
			cfg.JSONPrefix,
			cfg.JSONIndent,
			cfg.JSONFilePermission,
		); err != nil {
			fmt.Printf("Failed to output data: %v", err)
		}
	}
}