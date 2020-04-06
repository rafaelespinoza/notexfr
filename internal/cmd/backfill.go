package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/rafaelespinoza/snbackfill/internal/interactor"
)

var _Backfill = func(cmdName string) *delegator {
	cmd := &delegator{description: "synthesize resources among Evernote, StandardNotes"}
	cmd.subs = map[string]directive{
		"tags": &command{
			description: "find and match tags",
			setup: func(a *arguments) *flag.FlagSet {
				flags := setupBackfillSubcommand(a, cmdName+" tags")
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Tags,
					"input-en",
					"",
					"path to Evernote data file",
				)
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Tags,
					"output",
					"",
					"write output json to this file",
				)
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.MatchTags(
					context.TODO(),
					a.backfillOpts,
				)
				return err
			},
		},
		"notes": &command{
			description: "find and match notes",
			setup: func(a *arguments) *flag.FlagSet {
				flags := setupBackfillSubcommand(a, cmdName+" notes")
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Notes,
					"input-en",
					"",
					"path to Evernote data file",
				)
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Notes,
					"output",
					"",
					"write output json to this file",
				)
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.MatchNotes(
					context.TODO(),
					a.backfillOpts,
				)
				return err
			},
		},
		"notebooks": &command{
			description: "reconcile notebooks",
			setup: func(a *arguments) *flag.FlagSet {
				flags := setupBackfillSubcommand(a, cmdName+" notebooks")
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Notebooks,
					"input-en",
					"",
					"path to Evernote data file",
				)
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Notebooks,
					"output",
					"",
					"write output json to this file",
				)
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.ReconcileNotebooks(
					context.TODO(),
					a.backfillOpts,
				)
				return err
			},
		},
		"all": &command{
			description: "do it all",
			setup: func(a *arguments) *flag.FlagSet {
				flags := setupBackfillSubcommand(a, cmdName+" all")
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Notebooks,
					"input-en-notebooks",
					"",
					"path to Evernote notebooks data file",
				)
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Notes,
					"input-en-notes",
					"",
					"path to Evernote notes data file",
				)
				flags.StringVar(
					&a.backfillOpts.EvernoteFilenames.Tags,
					"input-en-tags",
					"",
					"path to Evernote tags data file",
				)
				flags.StringVar(
					&a.backfillOpts.OutputFilenames.Notebooks,
					"output-notebooks",
					"",
					"write notebooks json to this file",
				)
				flags.StringVar(
					&a.backfillOpts.OutputFilenames.Notes,
					"output-notes",
					"",
					"write notes json to this file",
				)
				flags.StringVar(
					&a.backfillOpts.OutputFilenames.Tags,
					"output-tags",
					"",
					"write tags json to this file",
				)
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.BackfillSN(
					context.TODO(),
					a.backfillOpts,
				)
				return err
			},
		},
	}
	descriptions := describeSubcommands(cmd.subs)
	cmd.flags = flag.NewFlagSet(cmdName, flag.ExitOnError)
	// TODO: briefly describe the input files and how to make them.
	cmd.flags.Usage = func() {
		mainsubName := _Bin + " " + cmdName
		fmt.Printf(`Usage: %s

Description:

	%s attempts to merge values in Evernote, StandardNotes resources.

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`, mainsubName, mainsubName, strings.Join(descriptions, "\n\t"),
		)
	}
	return cmd
}("backfill")

func setupBackfillSubcommand(a *arguments, name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	var opts interactor.BackfillOpts
	flags.StringVar(
		&opts.StandardNotesFilename,
		"input-sn",
		"",
		"path to StandardNotes data file",
	)
	flags.BoolVar(&opts.Verbose, "verbose", false, "output stuff as it happens")
	flags.Usage = func() {
		fmt.Printf(`Usage: %s %s`, _Bin, name)
		fmt.Printf("\n\nFlags:\n\n")
		flags.PrintDefaults()
	}
	a.backfillOpts = &opts
	return flags
}
