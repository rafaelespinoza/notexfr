package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
)

func makeBackfill(cmdName string) *cobra.Command {
	cmd := cobra.Command{
		Use:     cmdName,
		GroupID: dataGroupID,
		Short:   "supplement missing data for existing resources",
		Long: `Supplant metadata that may have been missed during an initial
transfer. Synthesize data from the source service into data that has already
been imported into the destination service.`,
	}
	enToSN := cobra.Command{
		Use:   "en-to-sn",
		Short: "backfill Evernote data for StandardNotes",
		Long: `Attempt to merge Evernote values into existing StandardNotes resources.

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

--input-en-notebooks=<output of "edam notebooks">
--input-en-notes=<output of "edam notes">
--input-en-tags=<output of "edam tags">

The input flag --input-sn is a StandardNotes export file. For example, the
one used to initally import your data from Evernote.

Results are written to new files where you can inspect them yourself.`,
	}
	{
		enToSN.Flags().StringP("input-sn", "", "", "path to StandardNotes data file")
		enToSN.Flags().StringP("input-en-notebooks", "", "", "path to Evernote notebooks data file")
		enToSN.Flags().StringP("input-en-notes", "", "", "path to Evernote notes data file")
		enToSN.Flags().StringP("input-en-tags", "", "", "path to Evernote tags data file")
		enToSN.Flags().StringP("output-notebooks", "", "", "write notebooks json to this file")
		enToSN.Flags().StringP("output-notes", "", "", "write notes json to this file")
		enToSN.Flags().StringP("output-tags", "", "", "write tags json to this file")

		enToSN.RunE = func(cmd *cobra.Command, args []string) error {
			var opts interactor.BackfillParams
			tuples := []struct {
				name string
				val  *string
			}{
				{name: "input-sn", val: &opts.StandardNotesFilename},
				{name: "input-en-notebooks", val: &opts.EvernoteFilenames.Notebooks},
				{name: "input-en-notes", val: &opts.EvernoteFilenames.Notes},
				{name: "input-en-tags", val: &opts.EvernoteFilenames.Tags},
				{name: "output-notebooks", val: &opts.OutputFilenames.Notebooks},
				{name: "output-notes", val: &opts.OutputFilenames.Notes},
				{name: "output-tags", val: &opts.OutputFilenames.Tags},
			}
			cmdFlags := cmd.Flags()
			for _, tuple := range tuples {
				val, err := cmdFlags.GetString(tuple.name)
				if err != nil {
					return err
				}
				*tuple.val = val
			}
			_, err := interactor.BackfillSN(cmd.Context(), &opts)
			return err
		}
	}

	cmd.AddCommand(&enToSN)
	return &cmd
}
