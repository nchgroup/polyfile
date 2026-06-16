package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"polyfile/internal/core/combiner"
	"polyfile/internal/core/parser"
)

var (
	buildOutput    string
	buildVerify    bool
	buildVerbose   bool
	buildStrategy  string
	buildFormat    string
	buildMagicDef  []string
	buildFormatDef []string
)

var buildCmd = &cobra.Command{
	Use:   "build <fileA> <fileB>",
	Short: "Combine two files into a polyglot output",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := parser.Parse(args, buildOutput, buildVerify)
		if err != nil {
			return err
		}

		if buildVerbose {
			fmt.Printf("pair:      %s + %s\n", cfg.Pair[0].ID(), cfg.Pair[1].ID())
			fmt.Printf("strategy:  %s\n", cfg.Strategy.Name())
			for i, inp := range cfg.Inputs {
				info, err := os.Stat(inp.Path)
				if err != nil {
					return err
				}
				fmt.Printf("input[%d]:  %-40s  %s  %d bytes\n",
					i, inp.Path, inp.Handler.ID(), info.Size())
			}
			fmt.Println()
		}

		result, err := combiner.Combine(cfg)
		if err != nil {
			return err
		}

		if buildVerbose {
			fmt.Printf("output:    %s\n", result.Path)
			fmt.Printf("size:      %d bytes\n", result.Size)
			fmt.Printf("sha256:    %s\n", result.SHA256)
		} else {
			fmt.Printf("output: %s\n", result.Path)
			fmt.Printf("size:   %d bytes\n", result.Size)
			fmt.Printf("sha256: %s\n", result.SHA256)
		}
		return nil
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output file path (required)")
	buildCmd.MarkFlagRequired("output")
	buildCmd.Flags().BoolVar(&buildVerify, "verify", true, "validate output after writing")
	buildCmd.Flags().BoolVarP(&buildVerbose, "verbose", "v", false, "print strategy, inputs, and offset summary")
	buildCmd.Flags().StringVar(&buildStrategy, "strategy", "", "force a specific combination strategy")
	buildCmd.Flags().StringVar(&buildFormat, "format", "", "explicit format pair override (e.g. pdf+zip)")
	buildCmd.Flags().StringArrayVar(&buildMagicDef, "magic-def", nil, "define custom format inline (repeatable)")
	buildCmd.Flags().StringArrayVar(&buildFormatDef, "format-def", nil, "load format definition from TOML file (repeatable)")
}
