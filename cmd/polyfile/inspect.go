package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"polyfile/internal/core/registry"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <file>",
	Short: "Show embedded formats and their offsets in a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return err
		}

		entries := registry.All()
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Handler.ID() < entries[j].Handler.ID()
		})

		fmt.Printf("file:  %s\nsize:  %d bytes\n\n", path, info.Size())
		fmt.Printf("%-10s  %s\n", "FORMAT", "STATUS")
		fmt.Printf("%-10s  %s\n", "------", "------")

		detected := 0
		for _, e := range entries {
			if err := e.Handler.Validate(f, info.Size()); err == nil {
				fmt.Printf("%-10s  valid\n", e.Handler.ID())
				detected++
			} else {
				fmt.Printf("%-10s  not detected (%v)\n", e.Handler.ID(), err)
			}
		}

		fmt.Printf("\n%d format(s) detected\n", detected)
		return nil
	},
}
