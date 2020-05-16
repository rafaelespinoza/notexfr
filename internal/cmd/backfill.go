package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
)

var _Backfill = func(cmdName string) *delegator {
	cmd := &delegator{description: "supplement missing data for existing resources"}
	cmd.subs = map[string]directive{
		"en-to-sn": &command{
			description: "backfill Evernote data for StandardNotes",
			setup: func(a *arguments) *flag.FlagSet {
				const subcmdName = " en-to-sn"
				flags := flag.NewFlagSet(cmdName+subcmdName, flag.ExitOnError)
				var opts interactor.BackfillParams
				flags.StringVar(&opts.StandardNotesFilename, "input-sn", "", "path to StandardNotes data file")
				flags.StringVar(&opts.EvernoteFilenames.Notebooks, "input-en-notebooks", "", "path to Evernote notebooks data file")
				flags.StringVar(&opts.EvernoteFilenames.Notes, "input-en-notes", "", "path to Evernote notes data file")
				flags.StringVar(&opts.EvernoteFilenames.Tags, "input-en-tags", "", "path to Evernote tags data file")
				flags.StringVar(&opts.OutputFilenames.Notebooks, "output-notebooks", "", "write notebooks json to this file")
				flags.StringVar(&opts.OutputFilenames.Notes, "output-notes", "", "write notes json to this file")
				flags.StringVar(&opts.OutputFilenames.Tags, "output-tags", "", "write tags json to this file")
				flags.BoolVar(&opts.Verbose, "verbose", false, "output stuff as it happens")
				flags.Usage = func() {
					fmt.Printf(`Usage: %s %s

	Attempt to merge Evernote values into existing StandardNotes resources.

	Use this command if you've already imported Evernote data into StandardNotes
	but you want to backfill certain metadata that was not initially transferred
	when using https://dashboard.standardnotes.org/tools. At the time of this
	writing, that tool uses an ENEX (Evernote export) file to construct
	StandardNotes data. Unfortunately, the ENEX format does not contain any
	Notebook info, so it can't be preserved with the existing conversion tool.
	You'd use this if you want to preserve your Evernote Notebooks and their
	Note associations.

	There are several input files to this subcommand. The following inputs are
	created from the edam subcommand. This is, you should fetch your data from
	Evernote using the EDAM API and save the results to local JSON files:

	--input-en-notebooks=<output of "%s edam notebooks">
	--input-en-notes=<output of "%s edam notes">
	--input-en-tags=<output of "%s edam tags">

	The input flag --input-sn is a StandardNotes export file. For example, the
	one used to initally import your data from Evernote.

	Results are written to new files where you can inspect them yourself.`,
						_Bin, subcmdName, _Bin, _Bin, _Bin)

					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				a.backfillOpts = &opts
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
	cmd.flags.Usage = func() {
		mainsubName := _Bin + " " + cmdName
		fmt.Printf(`Usage: %s

Description:

	%s attempts to supplant metadata that may have been missed during an initial
	transfer. Synthesize data from the source service into data that has already
	been imported into the destination service.

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`, mainsubName, mainsubName, strings.Join(descriptions, "\n\t"),
		)
	}
	return cmd
}("backfill")
