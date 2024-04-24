package cmd_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/rafaelespinoza/notexfr/internal/cmd"
)

const (
	_BaseTestOutputDir = "/tmp/notexfr_test/internal/cmd"
	_FixturesDir       = "../fixtures"
	_StubNotebooksFile = "edam_notebooks.json"
	_StubNotesFile     = "edam_notes.json"
	_StubTagsFile      = "edam_tags.json"
	_StubENEXFile      = "test_export.enex"
	_StubENtoSNFile    = "evernote-to-sn.txt"
)

func TestMain(m *testing.M) {
	if err := os.MkdirAll(_BaseTestOutputDir, 0700); err != nil {
		panic(err)
	}

	m.Run()
}

func TestBackfill(t *testing.T) {
	t.Run("en-to-sn", func(t *testing.T) {
		outputNotebooks := makeOutputFilenamePrefix(t) + "-notebooks.json"
		outputNotes := makeOutputFilenamePrefix(t) + "-notes.json"
		outputTags := makeOutputFilenamePrefix(t) + "-tags.json"
		os.Args = []string{
			"", "backfill", "en-to-sn",
			"--input-en-notebooks", _FixturesDir + "/" + _StubNotebooksFile,
			"--input-en-notes", _FixturesDir + "/" + _StubNotesFile,
			"--input-en-tags", _FixturesDir + "/" + _StubTagsFile,
			"--input-sn", _FixturesDir + "/" + _StubENtoSNFile,
			"--output-notebooks", outputNotebooks,
			"--output-notes", outputNotes,
			"--output-tags", outputTags,
		}
		cmd.Init()
		ctx := context.Background()
		err := cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("check outputs at %q", outputNotebooks)
		t.Logf("check outputs at %q", outputNotes)
		t.Logf("check outputs at %q", outputTags)
	})
}

func TestConvert(t *testing.T) {
	t.Run("edam-to-sn", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-output.json"
		os.Args = []string{
			"", "convert", "edam-to-sn",
			"--input-en-notebooks", _FixturesDir + "/" + _StubNotebooksFile,
			"--input-en-notes", _FixturesDir + "/" + _StubNotesFile,
			"--input-en-tags", _FixturesDir + "/" + _StubTagsFile,
			"--output", outputFilename,
		}
		cmd.Init()
		ctx := context.Background()
		err := cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("check output at %q", outputFilename)
	})

	t.Run("enex-to-sn", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-output.json"
		os.Args = []string{
			"", "convert", "enex-to-sn",
			"--input", _FixturesDir + "/" + _StubENEXFile,
			"--output", outputFilename,
		}
		cmd.Init()
		ctx := context.Background()
		err := cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("check output at %q", outputFilename)
	})
}

func TestEDAM(t *testing.T) {
	t.Run("make-env", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-envfile"
		defer os.Remove(outputFilename)

		os.Args = []string{
			"", "edam", "make-env",
			"-envfile", outputFilename,
		}
		cmd.Init()
		ctx := context.Background()
		err := cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		stat, err := os.Stat(outputFilename)
		if err != nil {
			t.Fatal(err)
		}
		if stat.Size() < 1 {
			t.Errorf("expected file %s to be non-empty", outputFilename)
		}
	})
}

func TestENEX(t *testing.T) {
	t.Run("inspect", func(t *testing.T) {
		os.Args = []string{
			"", "enex", "inspect",
			"--input", _FixturesDir + "/" + _StubENEXFile,
		}
		cmd.Init()
		ctx := context.Background()
		err := cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}

		os.Args = []string{
			"", "enex", "inspect",
			"--input", _FixturesDir + "/" + _StubENEXFile,
			"--pretty",
		}
		cmd.Init()
		err = cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("to-json", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-output.json"
		os.Args = []string{
			"", "enex", "to-json",
			"--input", _FixturesDir + "/" + _StubENEXFile,
			"--output", outputFilename,
		}
		cmd.Init()
		ctx := context.Background()
		err := cmd.Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("check output at %q", outputFilename)
	})
}

func TestVersion(t *testing.T) {
	os.Args = []string{"", "version"}
	cmd.Init()
	ctx := context.Background()
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func makeOutputFilenamePrefix(t *testing.T) string {
	return _BaseTestOutputDir + "/" + strings.Replace(t.Name(), "/", "_", -1)
}
