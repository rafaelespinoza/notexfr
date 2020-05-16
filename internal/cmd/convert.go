package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
)

var _Convert = func(cmdName string) *delegator {
	cmd := &delegator{description: "convert data"}
	cmd.subs = map[string]directive{
		"edam-to-sn": &command{
			description: "convert EDAM (Evernote) data to StandardNotes format",
			setup: func(a *arguments) *flag.FlagSet {
				subcmdName := cmdName + " edam-to-sn"
				var opts interactor.ConvertParams
				flags := flag.NewFlagSet(subcmdName, flag.ExitOnError)
				flags.StringVar(&opts.InputFilenames.Notebooks, "input-en-notebooks", "", "path to Evernote notebooks data file")
				flags.StringVar(&opts.InputFilenames.Notes, "input-en-notes", "", "path to Evernote notes data file")
				flags.StringVar(&opts.InputFilenames.Tags, "input-en-tags", "", "path to Evernote tags data file")
				flags.StringVar(&opts.OutputFilename, "output", "", "path to output file")
				flags.Usage = func() {
					fmt.Printf(`Usage: %s %s
	Parse, read local Evernote data, convert to StandardNotes JSON format.

	There are several input files to this subcommand. The following inputs are
	created from the edam subcommand. This is, you should fetch your data from
	Evernote using the EDAM API and save the results to local JSON files:

	--input-en-notebooks=<output of "%s edam notebooks">
	--input-en-notes=<output of "%s edam notes">
	--input-en-tags=<output of "%s edam tags">`,
						_Bin, subcmdName, _Bin, _Bin, _Bin)

					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				a.convertOpts = &opts
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.ConvertEDAMToStandardNotes(
					context.Background(),
					*a.convertOpts,
				)
				return err
			},
		},
		"enex-to-sn": &command{
			description: "convert an Evernote export file to StandardNotes format",
			setup: func(a *arguments) *flag.FlagSet {
				subcmdName := cmdName + " enex-to-sn"
				var opts interactor.ConvertParams
				flags := flag.NewFlagSet(subcmdName, flag.ExitOnError)
				flags.StringVar(&opts.InputFilename, "input", "", "path to evernote export file")
				flags.StringVar(&opts.OutputFilename, "output", "", "path to output file")
				flags.Usage = func() {
					fmt.Printf(`Usage: %s %s

	Parse, read an Evernote ENEX file, convert to StandardNotes JSON format.`,
						_Bin, subcmdName)

					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				a.convertOpts = &opts
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.ConvertENEXToStandardNotes(
					context.Background(),
					*a.convertOpts,
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

	%s converts data from one service format to another.

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`, mainsubName, mainsubName, strings.Join(descriptions, "\n\t"),
		)
	}
	return cmd
}("convert")
