package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"polyfile/internal/core/registry"
)

var formatsCmd = &cobra.Command{
	Use:   "formats",
	Short: "Manage and inspect the format registry",
}

var formatsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered formats",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries := registry.All()
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Handler.ID() < entries[j].Handler.ID()
		})

		fmt.Printf("%-12s %-20s %-10s %s\n", "ID", "EXTENSIONS", "SOURCE", "PATH")
		for _, e := range entries {
			ext := fmt.Sprintf("%v", e.Handler.Extensions())
			path := e.Path
			if path == "" {
				path = "-"
			}
			fmt.Printf("%-12s %-20s %-10s %s\n", e.Handler.ID(), ext, e.Source, path)
		}
		return nil
	},
}

var formatsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Print a format's full definition as TOML",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		h, err := registry.Lookup(args[0])
		if err != nil {
			return err
		}
		c := h.Constraints()
		fmt.Printf("id          = %q\n", h.ID())
		fmt.Printf("extensions  = %v\n", h.Extensions())
		fmt.Printf("\n[magic]\n")
		fmt.Printf("offset = %d\n", c.MagicOffset)
		fmt.Printf("bytes  = %X\n", c.Magic)
		fmt.Printf("\n[constraints]\n")
		fmt.Printf("can_prepend_arbitrary_bytes = %v\n", c.CanPrependArbitraryBytes)
		fmt.Printf("can_append_arbitrary_bytes  = %v\n", c.CanAppendArbitraryBytes)
		fmt.Printf("requires_trailer            = %v\n", c.RequiresTrailer)
		return nil
	},
}

var formatsValidateAs string

var formatsValidateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate magic bytes in a file against a known format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if formatsValidateAs == "" {
			return fmt.Errorf("--as <format-id> required")
		}
		fmt.Printf("validate %s as %s — not yet implemented\n", args[0], formatsValidateAs)
		return nil
	},
}

var formatsExportOut string

var formatsExportCmd = &cobra.Command{
	Use:   "export <id>",
	Short: "Export a built-in handler as a TOML template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("export %s — not yet implemented\n", args[0])
		return nil
	},
}

func init() {
	formatsValidateCmd.Flags().StringVar(&formatsValidateAs, "as", "", "format ID to validate against")
	formatsExportCmd.Flags().StringVar(&formatsExportOut, "out", "", "output TOML file path")

	formatsCmd.AddCommand(formatsListCmd)
	formatsCmd.AddCommand(formatsShowCmd)
	formatsCmd.AddCommand(formatsValidateCmd)
	formatsCmd.AddCommand(formatsExportCmd)
}
