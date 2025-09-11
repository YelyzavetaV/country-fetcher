package main

import (
	"fmt"

	"github.com/YelyzavetaV/country-fetcher/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Printf("Failed due to error: %v", err)
	}
}