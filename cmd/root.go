/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"todayfetch/fetcher"

	"github.com/spf13/cobra"
)

var (
	showEvents bool
	showEpoch  bool
	limit      int
	dbFile     = "todayfetch.db"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "todayfetch",
	Short: "A system fetch tool for today's date",
	Long:  `todayfetch displays interesting information about today, including historical events and UNIX time.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := fetcher.InitDB(dbFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer db.Close()

		fmt.Printf("--- Today is %s ---\n\n", fetcher.GetDate())

		// Default behavior: show everything if no flags are set
		allOff := !showEvents && !showEpoch

		if showEvents || allOff {
			fetcher.FetchEvents(db, limit)
		}

		if showEpoch || allOff {
			fetcher.FetchEpoch()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&showEvents, "events", "e", false, "Show historical events")
	rootCmd.Flags().BoolVarP(&showEpoch, "epoch", "x", false, "Show UNIX epoch")
	rootCmd.Flags().IntVarP(&limit, "limit", "l", 5, "Number of events to show")
}
