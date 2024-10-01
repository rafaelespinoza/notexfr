package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
	"github.com/rafaelespinoza/notexfr/internal/repo/enex"
)

func makeEnex(cmdName string) *cobra.Command {
	const helpLink = "https://help.evernote.com/hc/en-us/articles/209005557"

	cmd := cobra.Command{
		Use:     cmdName,
		GroupID: dataGroupID,
		Short:   "handle Evernote data via ENEX (Evernote export) files",
		Long: fmt.Sprintf(`Handles Evernote export data in the ENEX format.
For more info on exporting Evernote data, see: %s`, helpLink),
	}

	toJSON := cobra.Command{
		Use:   "to-json",
		Short: "convert ENEX file to JSON",
		Long: fmt.Sprintf(`Parse an Evernote export file and convert data to JSON entities.
For more info on exporting Evernote data, see: %s`, helpLink),
	}
	{
		toJSON.Flags().StringP("input", "i", "", "path to evernote export file")
		toJSON.Flags().StringP("output", "o", "", "path to write data as JSON")
		toJSON.Flags().DurationP("timeout", "t", 15*time.Second, "how long to wait before timing out")
		toJSON.Flags().BoolP("verbose", "v", false, "output stuff as it happens")

		toJSON.RunE = func(cmd *cobra.Command, args []string) (err error) {
			f := cmd.Flags()
			var params interactor.FetchWriteParams
			params.InputFilename, err = f.GetString("input")
			if err != nil {
				return
			}
			params.OutputFilename, err = f.GetString("output")
			if err != nil {
				return
			}
			params.Timeout, err = f.GetDuration("timeout")
			if err != nil {
				return
			}
			params.Verbose, err = f.GetBool("verbose")
			if err != nil {
				return
			}
			return interactor.WriteENEXToJSON(cmd.Context(), &params)
		}
	}

	inspect := cobra.Command{
		Use:   "inspect",
		Short: "inspect ENEX file as golang values",
		Long: fmt.Sprintf(`Parse, read an Evernote ENEX file, convert to golang values, output to stdout.
Could be useful for debugging or inspecting data in development.
For more info on exporting Evernote data, see: %s`, helpLink),
	}
	{
		inspect.Flags().StringP("input", "i", "", "path to evernote export file")
		inspect.Flags().BoolP("pretty", "p", false, "pretty print output")
		inspect.RunE = func(cmd *cobra.Command, args []string) (err error) {
			var opts enex.FileOpts
			f := cmd.Flags()
			opts.Filename, err = f.GetString("input")
			if err != nil {
				return
			}
			opts.PrettyPrint, err = f.GetBool("pretty")
			if err != nil {
				return
			}
			return enex.ReadPrintFile(cmd.Context(), &opts)
		}
	}

	cmd.AddCommand(&toJSON, &inspect)

	return &cmd
}
