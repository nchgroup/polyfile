package main

import (
	"github.com/spf13/cobra"

	// Register built-in format handlers via init().
	_ "polyfile/internal/formats/gif"
	_ "polyfile/internal/formats/jpg"
	_ "polyfile/internal/formats/mp3"
	_ "polyfile/internal/formats/pdf"
	_ "polyfile/internal/formats/png"
	_ "polyfile/internal/formats/sh"
	_ "polyfile/internal/formats/zip"
)

var rootCmd = &cobra.Command{
	Use:   "polyfile",
	Short: "Create polyglot files valid in two or more formats simultaneously",
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(formatsCmd)
	rootCmd.AddCommand(versionCmd)
}
