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
		args := []string{
			"backfill", "en-to-sn",
			"--input-en-notebooks", _FixturesDir + "/" + _StubNotebooksFile,
			"--input-en-notes", _FixturesDir + "/" + _StubNotesFile,
			"--input-en-tags", _FixturesDir + "/" + _StubTagsFile,
			"--input-sn", _FixturesDir + "/" + _StubENtoSNFile,
			"--output-notebooks", outputNotebooks,
			"--output-notes", outputNotes,
			"--output-tags", outputTags,
		}
		runOrDie(t, args)
		t.Logf("check outputs at %q", outputNotebooks)
		t.Logf("check outputs at %q", outputNotes)
		t.Logf("check outputs at %q", outputTags)
	})
}

func TestConvert(t *testing.T) {
	t.Run("edam-to-sn", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-output.json"
		args := []string{
			"convert", "edam-to-sn",
			"--input-en-notebooks", _FixturesDir + "/" + _StubNotebooksFile,
			"--input-en-notes", _FixturesDir + "/" + _StubNotesFile,
			"--input-en-tags", _FixturesDir + "/" + _StubTagsFile,
			"--output", outputFilename,
		}
		runOrDie(t, args)
		t.Logf("check output at %q", outputFilename)
	})

	t.Run("enex-to-sn", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-output.json"
		args := []string{
			"convert", "enex-to-sn",
			"--input", _FixturesDir + "/" + _StubENEXFile,
			"--output", outputFilename,
		}
		runOrDie(t, args)
		t.Logf("check output at %q", outputFilename)
	})
}

func TestEDAM(t *testing.T) {
	t.Run("make-env", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-envfile"
		defer os.Remove(outputFilename)

		args := []string{
			"edam", "make-env",
			"--envfile", outputFilename,
		}
		runOrDie(t, args)
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
		args := []string{
			"enex", "inspect",
			"--input", _FixturesDir + "/" + _StubENEXFile,
		}
		runOrDie(t, args)

		args = []string{
			"enex", "inspect",
			"--input", _FixturesDir + "/" + _StubENEXFile,
			"--pretty",
		}
		runOrDie(t, args)
	})

	t.Run("to-json", func(t *testing.T) {
		outputFilename := makeOutputFilenamePrefix(t) + "-output.json"
		args := []string{
			"enex", "to-json",
			"--input", _FixturesDir + "/" + _StubENEXFile,
			"--output", outputFilename,
		}
		runOrDie(t, args)
		t.Logf("check output at %q", outputFilename)
	})
}

func TestVersion(t *testing.T) {
	runOrDie(t, []string{"version"})
}

func makeOutputFilenamePrefix(t *testing.T) string {
	return _BaseTestOutputDir + "/" + strings.Replace(t.Name(), "/", "_", -1)
}

func runOrDie(t *testing.T, args []string) {
	t.Helper()

	root := cmd.New()
	os.Args = append([]string{""}, args...)

	err := root.ExecuteContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
