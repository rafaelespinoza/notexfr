package interactor_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/rafaelespinoza/snbackfill/internal/entity"
	"github.com/rafaelespinoza/snbackfill/internal/interactor"
	"github.com/rafaelespinoza/snbackfill/internal/repo/sn"
)

const (
	_BaseTestOutputDir = "/tmp/snbackfill_test/entity/interactor"
	_FixturesDir       = "../../internal/fixtures"
	_StubNotebooksFile = "edam_notebooks.json"
	_StubNotesFile     = "edam_notes.json"
	_StubTagsFile      = "edam_tags.json"
	_StubENtoSNFile    = "evernote-to-sn.txt"
)

func TestMain(m *testing.M) {
	os.MkdirAll(_BaseTestOutputDir, 0755)
	m.Run()
	os.RemoveAll(_BaseTestOutputDir)
}

func TestInterfaceImplementations(t *testing.T) {
	t.Run("members", func(t *testing.T) {
		implementations := []interface{}{
			new(interactor.FromENToSN),
		}
		for i, val := range implementations {
			var ok bool
			if _, ok = val.(entity.LinkID); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement entity.LinkID",
					i, val,
				)
			}
		}
	})
}

func TestReconciliation(t *testing.T) {
	pathToTestDir := _BaseTestOutputDir + "/" + t.Name()
	if err := os.MkdirAll(pathToTestDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(pathToTestDir)
	// expectedTestValues is a set of expected outputs for one test case.
	type expectedTestValues struct {
		SNID      string
		ENID      string
		Title     string
		CreatedAt time.Time
		UpdatedAt time.Time
		TagIDs    []string
	}

	t.Run("Tags", func(t *testing.T) {
		var (
			actualTags []entity.LinkID
			err        error
			conv       *interactor.FromENToSN
			tag        *sn.Tag
			ok         bool
		)
		actualTags, err = interactor.MatchTags(
			context.Background(),
			&interactor.BackfillOpts{
				EvernoteFilenames: struct{ Notebooks, Notes, Tags string }{
					Tags: _FixturesDir + "/" + _StubTagsFile,
				},
				StandardNotesFilename: _FixturesDir + "/" + _StubENtoSNFile,
				OutputFilenames: struct{ Notebooks, Notes, Tags string }{
					Tags: pathToTestDir + "/tags.json",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := []expectedTestValues{
			{
				SNID:  "845e98ed-4515-473e-836b-ada5b5cb8d01",
				ENID:  "ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
				Title: "foo",
			},
			{
				SNID:  "4cbfa0b5-655c-447c-9961-6a5294b6b041",
				ENID:  "53f1fdb5-4140-4ff4-8590-21e8cc2b4338",
				Title: "baker",
			},
			{
				SNID:  "e5daa664-db99-4cbf-afe5-2c0f043bac8c",
				ENID:  "f299e07c-b98e-4902-ac69-d0c7927e4870",
				Title: "free",
			},
			{
				SNID:  "30cf9510-845d-4ea8-b673-51104a3e0bc2",
				ENID:  "bb170464-72e0-4a22-85e0-b2c4f68272ea",
				Title: "bar",
			},
			// This is in the conversion file, but not in the edam file.
			// TODO: figure out what to do.
			// {
			// 	SNID:  "90e045a2-46ea-44ff-808b-648274926c7f",
			// 	Title: "evernote",
			// },
		}
		if len(actualTags) != len(expectedOutput) {
			t.Errorf(
				"wrong output length; got %d, expected %d",
				len(actualTags), len(expectedOutput),
			)
		}

		for i, resource := range actualTags {
			if resource.GetID() != expectedOutput[i].SNID {
				t.Errorf(
					"test %d; wrong ID; got %q, expected %q",
					i, resource.GetID(), expectedOutput[i].SNID,
				)
			}
			if conv, ok = resource.(*interactor.FromENToSN); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, resource, &interactor.FromENToSN{},
				)
			}
			if conv.EvernoteID.GetID() != expectedOutput[i].ENID {
				t.Errorf(
					"test %d; wrong ENID; got %q, expected %q",
					i, conv.EvernoteID.GetID(), expectedOutput[i].ENID,
				)
			}
			if tag, ok = conv.LinkID.(*sn.Tag); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, resource, &sn.Tag{},
				)
			}
			if tag.Content.Title != expectedOutput[i].Title {
				t.Errorf(
					"test %d; wrong Title; got %q, expected %q",
					i, tag.Content.Title, expectedOutput[i].Title,
				)
			}
			// We don't care about timestamp values here because the EDAM API does
			// not have it for Tag and the value in the StandardNote conversion
			// files is the time it was converted, which is not relevant.
		}
	})

	t.Run("Notes", func(t *testing.T) {
		var (
			actualOutput []entity.LinkID
			err          error
			conv         *interactor.FromENToSN
			note         *sn.Note
			ok           bool
		)
		actualOutput, err = interactor.MatchNotes(
			context.Background(),
			&interactor.BackfillOpts{
				EvernoteFilenames: struct{ Notebooks, Notes, Tags string }{
					Notes: _FixturesDir + "/" + _StubNotesFile,
				},
				StandardNotesFilename: _FixturesDir + "/" + _StubENtoSNFile,
				OutputFilenames: struct{ Notebooks, Notes, Tags string }{
					Notes: pathToTestDir + "/notes.json",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		expectedOutput := []expectedTestValues{
			{
				SNID:      "8e053669-d1cc-4b69-a7fd-4433fc48feb7",
				ENID:      "8c44eeb1-7e50-4edb-95c4-12cf90d1017e",
				CreatedAt: time.Date(2020, 3, 7, 20, 21, 56, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 25, 54, 0, time.UTC),
				Title:     "Batman",
			},
			{
				SNID:      "8154cd4e-dd06-4386-afe4-ec09f847b708",
				ENID:      "7c56e278-c268-4003-b1cd-09853ad92b4a",
				CreatedAt: time.Date(2020, 3, 7, 20, 22, 13, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 33, 33, 0, time.UTC),
				Title:     "Atlanta",
			},
			{
				SNID:      "228e48b8-1f46-4c79-a429-a093ed21656c",
				ENID:      "1820018f-1d5e-4ae9-92c3-f0d72f45d25c",
				CreatedAt: time.Date(2020, 3, 8, 22, 18, 21, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 8, 22, 18, 36, 0, time.UTC),
				Title:     "Hello World",
			},
			{
				SNID:      "148dbae4-14b7-420e-8cb0-448d1be90ec6",
				ENID:      "4a5704f8-0825-4926-8f9e-ca74b3c7da85",
				CreatedAt: time.Date(2020, 3, 7, 20, 26, 3, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 26, 9, 0, time.UTC),
				Title:     "Chicago",
			},
			{
				SNID:      "7d43e417-d6f2-4a7e-9927-60d914c75a45",
				ENID:      "25480cfd-5785-4741-a6fd-a3e37aa9d43e",
				CreatedAt: time.Date(2020, 3, 7, 20, 33, 13, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 33, 21, 0, time.UTC),
				Title:     "Edmonton",
			},
			{
				SNID:      "16245526-2f33-4024-ab55-f6c17a61c053",
				ENID:      "9901a8a3-39a6-437c-9443-e60ef83e6394",
				CreatedAt: time.Date(2020, 3, 7, 20, 25, 24, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 25, 29, 0, time.UTC),
				Title:     "Baltimore",
			},
			{
				SNID:      "9c78d572-bf80-4fd5-98ce-5b21624100fe",
				ENID:      "879f5e58-60aa-496b-b764-bee8cfd664f6",
				CreatedAt: time.Date(2020, 3, 7, 20, 33, 36, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 34, 48, 0, time.UTC),
				Title:     "Fargo",
			},
			{
				SNID:      "27822e0b-7e2e-4ef3-ab8b-164f7932abfb",
				ENID:      "04630bf8-0800-408b-97d8-cebba0e8b864",
				CreatedAt: time.Date(2020, 3, 7, 20, 29, 39, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 29, 44, 0, time.UTC),
				Title:     "Despicable Me",
			},
			{
				SNID:      "3d378ee3-774f-49cc-838d-4fec925c593a",
				ENID:      "e0322fce-4633-4d7d-8dff-79664844f03f",
				CreatedAt: time.Date(2020, 3, 7, 20, 30, 42, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 30, 50, 0, time.UTC),
				Title:     "Fargo",
			},
			{
				SNID:      "581a0e66-2cbf-4976-b5e2-a5c6705ca0af",
				ENID:      "c66bca64-4395-4675-ae86-9ef35cc0e5cf",
				CreatedAt: time.Date(2020, 3, 7, 20, 30, 5, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 30, 9, 0, time.UTC),
				Title:     "Enter The Dragon",
			},
			{
				SNID:      "59f60f6f-fa97-4eaa-80a3-0006e345d381",
				ENID:      "a2197031-1570-40e4-bc8f-0cb776057f6b",
				CreatedAt: time.Date(2020, 3, 7, 20, 32, 37, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 32, 43, 0, time.UTC),
				Title:     "Detroit",
			},
			{
				SNID:      "0f4fe851-b6ff-455a-8d10-cc2e441f7deb",
				ENID:      "9df1e45a-623d-4e1b-b370-2d2365499ed0",
				CreatedAt: time.Date(2020, 3, 7, 20, 25, 10, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 25, 16, 0, time.UTC),
				Title:     "Casino",
			},
			{
				SNID:      "63f6d85b-1ac8-4ccf-8ae7-58007fcd033d",
				ENID:      "0f32f51b-f923-4dd0-b5c3-a7d6e8a8a40f",
				CreatedAt: time.Date(2020, 3, 7, 20, 24, 3, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 24, 47, 0, time.UTC),
				Title:     "Aladdin",
			},
		}
		if len(actualOutput) != len(expectedOutput) {
			t.Errorf(
				"wrong output length; got %d, expected %d",
				len(actualOutput), len(expectedOutput),
			)
		}

		for i, item := range actualOutput {
			if item.GetID() != expectedOutput[i].SNID {
				t.Errorf(
					"test %d; wrong ID; got %q, expected %q",
					i, item.GetID(), expectedOutput[i].SNID,
				)
			}
			if conv, ok = item.(*interactor.FromENToSN); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, item, &interactor.FromENToSN{},
				)
			}
			if conv.EvernoteID.GetID() != expectedOutput[i].ENID {
				t.Errorf(
					"test %d; wrong ENID; got %q, expected %q",
					i, conv.EvernoteID.GetID(), expectedOutput[i].ENID,
				)
			}
			if note, ok = conv.LinkID.(*sn.Note); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, item, &sn.Tag{},
				)
			}
			if note.Content.Title != expectedOutput[i].Title {
				t.Errorf(
					"test %d; wrong Title; got %q, expected %q",
					i, note.Content.Title, expectedOutput[i].Title,
				)
			}
			if !note.CreatedAt.Equal(expectedOutput[i].CreatedAt) {
				t.Errorf(
					"test %d; wrong CreatedAt; got %q, expected %q",
					i, note.CreatedAt, expectedOutput[i].CreatedAt,
				)
			}
			if !note.UpdatedAt.Equal(expectedOutput[i].UpdatedAt) {
				t.Errorf(
					"test %d; wrong UpdatedAt; got %q, expected %q",
					i, note.UpdatedAt, expectedOutput[i].UpdatedAt,
				)
			}
		}
	})

	t.Run("Notebooks", func(t *testing.T) {
		var (
			actualOutput []entity.LinkID
			err          error
			conv         *interactor.FromENToSN
			tag          *sn.Tag
			ok           bool
		)
		actualOutput, err = interactor.ReconcileNotebooks(
			context.Background(),
			&interactor.BackfillOpts{
				EvernoteFilenames: struct{ Notebooks, Notes, Tags string }{
					Notebooks: _FixturesDir + "/" + _StubNotebooksFile,
				},
				StandardNotesFilename: _FixturesDir + "/" + _StubENtoSNFile,
				OutputFilenames: struct{ Notebooks, Notes, Tags string }{
					Notebooks: pathToTestDir + "/notebooks.json",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// This output built from a map, which does not have a guaranteed
		// iteration order. To simplify tests, we need to compare items in the
		// same order. Sorting is not an essential feature in the source code
		// right now, so we sort only where needed.
		sort.Slice(actualOutput, func(i, j int) bool {
			left, right := mustSNTag(actualOutput[i]), mustSNTag(actualOutput[j])
			if left.Content.Title < right.Content.Title {
				return true
			} else if left.Content.Title > right.Content.Title {
				return false
			} else {
				return left.UUID < right.UUID
			}
		})

		expectedOutput := []expectedTestValues{
			{
				ENID:      "6aac8caa-0682-4870-95fb-f384301704bc",
				CreatedAt: time.Date(2020, 3, 7, 2, 13, 22, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 8, 22, 17, 53, 0, time.UTC),
				Title:     "<Inbox>",
			},
			{
				ENID:      "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				CreatedAt: time.Date(2020, 3, 7, 2, 14, 49, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 23, 35, 0, time.UTC),
				Title:     "Cities",
			},
			{
				ENID:      "932d7c12-bb87-4b41-895a-5d30fa178688",
				CreatedAt: time.Date(2020, 3, 7, 2, 14, 58, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 23, 55, 0, time.UTC),
				Title:     "Movies",
			},
			{
				Title: "SampleStack Stack",
			},
			{
				ENID:      "ca62b4e6-5649-4512-ae30-2a8de03f80fe",
				CreatedAt: time.Date(2020, 3, 8, 22, 16, 20, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 8, 22, 44, 6, 0, time.UTC),
				Title:     "Samples",
			},
		}
		if len(actualOutput) != len(expectedOutput) {
			t.Errorf(
				"wrong output length; got %d, expected %d",
				len(actualOutput), len(expectedOutput),
			)
		}

		for i, item := range actualOutput {
			if conv, ok = item.(*interactor.FromENToSN); !ok {
				t.Fatalf(
					"test %d; wrong type; expected %T",
					i, &interactor.FromENToSN{},
				)
			}
			if conv.EvernoteID.GetID() != expectedOutput[i].ENID {
				t.Errorf(
					"test %d; wrong ENID; got %q, expected %q",
					i, conv.EvernoteID.GetID(), expectedOutput[i].ENID,
				)
			}
			if tag, ok = conv.LinkID.(*sn.Tag); !ok {
				t.Fatalf(
					"test %d; wrong type for LinkID field; expected %T; entire item: %#v",
					i, &sn.Tag{}, item,
				)
			}
			if tag.Content.Title != expectedOutput[i].Title {
				t.Errorf(
					"test %d; wrong Title; got %q, expected %q",
					i, tag.Content.Title, expectedOutput[i].Title,
				)
			}
			if !tag.CreatedAt.Equal(expectedOutput[i].CreatedAt) {
				t.Errorf(
					"test %d; wrong CreatedAt; got %q, expected %q",
					i, tag.CreatedAt, expectedOutput[i].CreatedAt,
				)
			}
			if !tag.UpdatedAt.Equal(expectedOutput[i].UpdatedAt) {
				t.Errorf(
					"test %d; wrong UpdatedAt; got %q, expected %q",
					i, tag.UpdatedAt, expectedOutput[i].UpdatedAt,
				)
			}
		}
	})

	t.Run("BackfillSN", func(t *testing.T) {
		var (
			actualOutput []entity.LinkID
			err          error
			conv         *interactor.FromENToSN
			note         *sn.Note
			ok           bool
		)
		actualOutput, err = interactor.BackfillSN(
			context.TODO(),
			&interactor.BackfillOpts{
				EvernoteFilenames: struct{ Notebooks, Notes, Tags string }{
					Notebooks: _FixturesDir + "/" + _StubNotebooksFile,
					Notes:     _FixturesDir + "/" + _StubNotesFile,
					Tags:      _FixturesDir + "/" + _StubTagsFile,
				},
				StandardNotesFilename: _FixturesDir + "/" + _StubENtoSNFile,
				OutputFilenames: struct{ Notebooks, Notes, Tags string }{
					Notes: pathToTestDir + "/all_the_things.json",
				},
				Verbose: false,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		sort.Slice(actualOutput, func(i, j int) bool {
			left, right := mustSNNote(actualOutput[i]), mustSNNote(actualOutput[j])
			if left.Content.Title < right.Content.Title {
				return true
			} else if left.Content.Title > right.Content.Title {
				return false
			} else {
				return left.UUID < right.UUID
			}
		})

		knownNotebookIDs := map[string]string{
			"Cities":  "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
			"Movies":  "932d7c12-bb87-4b41-895a-5d30fa178688",
			"<Inbox>": "6aac8caa-0682-4870-95fb-f384301704bc",
			"Samples": "ca62b4e6-5649-4512-ae30-2a8de03f80fe",
		}
		knownTagIDs := map[string]string{
			"baker":    "4cbfa0b5-655c-447c-9961-6a5294b6b041",
			"bar":      "30cf9510-845d-4ea8-b673-51104a3e0bc2",
			"evernote": "90e045a2-46ea-44ff-808b-648274926c7f",
			"foo":      "845e98ed-4515-473e-836b-ada5b5cb8d01",
			"free":     "e5daa664-db99-4cbf-afe5-2c0f043bac8c",
		}

		expectedOutput := []expectedTestValues{
			{
				SNID:  "63f6d85b-1ac8-4ccf-8ae7-58007fcd033d",
				ENID:  "0f32f51b-f923-4dd0-b5c3-a7d6e8a8a40f",
				Title: "Aladdin",
				TagIDs: []string{
					knownTagIDs["bar"],
					knownNotebookIDs["Movies"],
				},
			},
			{
				SNID:  "8154cd4e-dd06-4386-afe4-ec09f847b708",
				ENID:  "7c56e278-c268-4003-b1cd-09853ad92b4a",
				Title: "Atlanta",
				TagIDs: []string{
					knownTagIDs["foo"],
					knownNotebookIDs["Cities"],
				},
			},
			{
				SNID:  "16245526-2f33-4024-ab55-f6c17a61c053",
				ENID:  "9901a8a3-39a6-437c-9443-e60ef83e6394",
				Title: "Baltimore",
				TagIDs: []string{
					knownTagIDs["bar"],
					knownNotebookIDs["Cities"],
				},
			},
			{
				SNID:  "8e053669-d1cc-4b69-a7fd-4433fc48feb7",
				ENID:  "8c44eeb1-7e50-4edb-95c4-12cf90d1017e",
				Title: "Batman",
				TagIDs: []string{
					knownTagIDs["foo"],
					knownTagIDs["baker"],
					knownNotebookIDs["Movies"],
				},
			},
			{
				SNID:  "0f4fe851-b6ff-455a-8d10-cc2e441f7deb",
				ENID:  "9df1e45a-623d-4e1b-b370-2d2365499ed0",
				Title: "Casino",
				TagIDs: []string{
					knownTagIDs["bar"],
					knownNotebookIDs["Movies"],
				},
			},
			{
				SNID:  "148dbae4-14b7-420e-8cb0-448d1be90ec6",
				ENID:  "4a5704f8-0825-4926-8f9e-ca74b3c7da85",
				Title: "Chicago",
				TagIDs: []string{
					knownTagIDs["baker"],
					knownTagIDs["free"],
					knownNotebookIDs["Cities"],
				},
			},
			{
				SNID:  "27822e0b-7e2e-4ef3-ab8b-164f7932abfb",
				ENID:  "04630bf8-0800-408b-97d8-cebba0e8b864",
				Title: "Despicable Me",
				TagIDs: []string{
					knownTagIDs["foo"],
					knownNotebookIDs["Movies"],
				},
			},
			{
				SNID:  "59f60f6f-fa97-4eaa-80a3-0006e345d381",
				ENID:  "a2197031-1570-40e4-bc8f-0cb776057f6b",
				Title: "Detroit",
				TagIDs: []string{
					knownTagIDs["bar"],
					knownNotebookIDs["Cities"],
				},
			},
			{
				SNID:  "7d43e417-d6f2-4a7e-9927-60d914c75a45",
				ENID:  "25480cfd-5785-4741-a6fd-a3e37aa9d43e",
				Title: "Edmonton",
				TagIDs: []string{
					knownTagIDs["foo"],
					knownNotebookIDs["Cities"],
				},
			},
			{
				SNID:  "581a0e66-2cbf-4976-b5e2-a5c6705ca0af",
				ENID:  "c66bca64-4395-4675-ae86-9ef35cc0e5cf",
				Title: "Enter The Dragon",
				TagIDs: []string{
					knownTagIDs["bar"],
					knownNotebookIDs["Movies"],
				},
			},
			{
				SNID:  "3d378ee3-774f-49cc-838d-4fec925c593a",
				ENID:  "e0322fce-4633-4d7d-8dff-79664844f03f",
				Title: "Fargo",
				TagIDs: []string{
					knownTagIDs["foo"],
					knownNotebookIDs["Movies"],
				},
			},
			{
				SNID:  "9c78d572-bf80-4fd5-98ce-5b21624100fe",
				ENID:  "879f5e58-60aa-496b-b764-bee8cfd664f6",
				Title: "Fargo",
				TagIDs: []string{
					knownTagIDs["bar"],
					knownNotebookIDs["Cities"],
				},
			},
			{
				SNID:  "228e48b8-1f46-4c79-a429-a093ed21656c",
				ENID:  "1820018f-1d5e-4ae9-92c3-f0d72f45d25c",
				Title: "Hello World",
				TagIDs: []string{
					knownTagIDs["foo"],
					knownTagIDs["baker"],
					knownNotebookIDs["<Inbox>"],
				},
			},
		}

		if len(actualOutput) != len(expectedOutput) {
			t.Errorf(
				"wrong output length; got %d, expected %d",
				len(actualOutput), len(expectedOutput),
			)
		}

		for i, item := range actualOutput {
			if item.GetID() != expectedOutput[i].SNID {
				t.Errorf(
					"test %d; wrong ID; got %q, expected %q",
					i, item.GetID(), expectedOutput[i].SNID,
				)
			}
			if conv, ok = item.(*interactor.FromENToSN); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, item, &interactor.FromENToSN{},
				)
			}
			if conv.EvernoteID.GetID() != expectedOutput[i].ENID {
				t.Errorf(
					"test %d; wrong ENID; got %q, expected %q",
					i, conv.EvernoteID.GetID(), expectedOutput[i].ENID,
				)
			}
			if note, ok = conv.LinkID.(*sn.Note); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, item, &sn.Tag{},
				)
			}
			if note.Content.Title != expectedOutput[i].Title {
				t.Errorf(
					"test %d; wrong Title; got %q, expected %q",
					i, note.Content.Title, expectedOutput[i].Title,
				)
			}
			if len(note.Content.References) != len(expectedOutput[i].TagIDs) {
				t.Errorf(
					"test %d; wrong number of tag IDs; got %d, expected %d",
					i, len(note.Content.References), len(expectedOutput[i].TagIDs),
				)
			}
			for j, ref := range note.Content.References {
				if ref.UUID != expectedOutput[i].TagIDs[j] {
					t.Errorf(
						"test [%d][%d] wrong tagID; got %q, expected %q",
						i, j, ref.UUID, expectedOutput[i].TagIDs[j],
					)
				}
			}
		}
	})
}

func mustSNTag(link entity.LinkID) *sn.Tag {
	item := link.(*interactor.FromENToSN)
	return item.LinkID.(*sn.Tag)
}

func mustSNNote(link entity.LinkID) *sn.Note {
	item := link.(*interactor.FromENToSN)
	return item.LinkID.(*sn.Note)
}

func mustPrettyPrintJSON(in interface{}) {
	out, err := json.MarshalIndent(in, "  ", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", string(out))
}
