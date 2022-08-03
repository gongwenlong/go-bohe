package main

import (
	"github.com/gongwenlong/go-bohe/cmd/bohe/internal/run"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "bohe",
	Short:   "bohe: An elegant toolkit for Go microservices.",
	Long:    `bohe: An elegant toolkit for Go microservices.`,
	Version: release,
}

func init() {
	rootCmd.AddCommand(run.CmdRun)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
