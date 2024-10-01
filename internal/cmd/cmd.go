// Package cmd wraps up all command line interface operations.
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

const (
	mainName    = "notexfr"
	dataGroupID = "data_group"
)

// Root abstracts a top-level command from package main.
type Root interface {
	// ExecuteContext is the entry point.
	ExecuteContext(ctx context.Context) error
}

// New establishes the root comand and its subcommands.
func New() Root {
	out := cobra.Command{
		Use:   mainName,
		Short: "main command for " + mainName,
		Long: `A tool for converting note data to other service formats.

Currently supported services
- Evernote
- StandardNotes`,
	}

	out.AddCommand(
		makeBackfill("backfill"),
		makeConvert("convert"),
		makeEdam("edam"),
		makeEnex("enex"),
		makeVersion("version"),
	)
	out.AddGroup(
		&cobra.Group{ID: dataGroupID, Title: "Data Commands:"},
	)
	return &out
}
