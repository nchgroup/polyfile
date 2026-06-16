package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"polyfile/internal/core/registry"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported format pairs",
	RunE: func(cmd *cobra.Command, args []string) error {
		pairs := registry.AllPairs()
		if len(pairs) == 0 {
			fmt.Println("no format pairs registered")
			return nil
		}

		fmt.Printf("%-12s  %-16s  %-18s  %s\n", "PAIR", "FORMAT A", "STRATEGY", "NOTES")
		fmt.Printf("%-12s  %-16s  %-18s  %s\n",
			"------------", "----------------", "------------------", "-----")
		for _, p := range pairs {
			pair := p.A + "+" + p.B
			fmt.Printf("%-12s  %-8s+%-7s  %-18s  %s\n",
				pair, p.A, p.B, p.Strategy, p.Notes)
		}
		return nil
	},
}
