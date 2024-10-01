package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
)

func makeConvert(cmdName string) *cobra.Command {
	cmd := cobra.Command{
		Use:     cmdName,
		GroupID: dataGroupID,
		Short:   "convert data",
		Long:    "converts data from one service format to another",
	}

	edamToSN := cobra.Command{
		Use:   "edam-to-sn",
		Short: "convert EDAM (Evernote) data to StandardNotes format",
		Long: `Parse, read local Evernote data, convert to StandardNotes JSON format.

There are several input files to this subcommand. The following inputs are
created from the edam subcommand. This is, you should fetch your data from
Evernote using the EDAM API and save the results to local JSON files:

--input-en-notebooks=<output of "edam notebooks">
--input-en-notes=<output of "edam notes">
--input-en-tags=<output of "edam tags">`,
	}
	{
		edamToSN.Flags().StringP("input-en-notebooks", "", "", "path to Evernote notebooks data file")
		edamToSN.Flags().StringP("input-en-notes", "", "", "path to Evernote notes data file")
		edamToSN.Flags().StringP("input-en-tags", "", "", "path to Evernote tags data file")
		edamToSN.Flags().StringP("output", "o", "", "path to output file")
		edamToSN.RunE = func(cmd *cobra.Command, args []string) (err error) {
			flags := cmd.Flags()
			var params interactor.ConvertParams
			params.InputFilenames.Notebooks, err = flags.GetString("input-en-notebooks")
			if err != nil {
				return err
			}
			params.InputFilenames.Notes, err = flags.GetString("input-en-notes")
			if err != nil {
				return err
			}
			params.InputFilenames.Tags, err = flags.GetString("input-en-tags")
			if err != nil {
				return err
			}
			params.OutputFilename, err = flags.GetString("output")
			if err != nil {
				return err
			}

			_, err = interactor.ConvertEDAMToStandardNotes(cmd.Context(), params)
			return err
		}
	}

	enexToSN := cobra.Command{
		Use:   "enex-to-sn",
		Short: "convert an Evernote export file to StandardNotes format",
		Long:  "Parse, read an Evernote ENEX file, convert to StandardNotes JSON format",
	}
	{
		enexToSN.Flags().StringP("input", "i", "", "path to evernote export file")
		enexToSN.Flags().StringP("output", "o", "", "path to output file")
		enexToSN.RunE = func(cmd *cobra.Command, args []string) (err error) {
			flags := cmd.Flags()
			var params interactor.ConvertParams
			params.InputFilename, err = flags.GetString("input")
			if err != nil {
				return err
			}
			params.OutputFilename, err = flags.GetString("output")
			if err != nil {
				return err
			}

			_, err = interactor.ConvertENEXToStandardNotes(cmd.Context(), params)
			return err
		}
	}

	cmd.AddCommand(&edamToSN, &enexToSN)
	return &cmd
}
