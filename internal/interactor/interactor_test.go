package interactor_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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

func TestConvert(t *testing.T) {
	pathToTestDir := _BaseTestOutputDir + "/" + t.Name()
	if err := os.MkdirAll(pathToTestDir, 0755); err != nil {
		t.Fatal(err)
	}

	type TestReference struct {
		// Type is a human-readable label for the kind of data referenced. It
		// should be one of: "Note", "Tag", "Notebook".
		Type string
		// Index should be the slice index of a referenced item.
		Index int
	}

	// ExpectedTestValues is a set of expectations for one item in the output.
	type ExpectedTestValues struct {
		ContentType string
		Title       string
		// ExpectNoteText signals that the we expect this item to have some
		// captured HTML content. This would only be true for Note items in the
		// test input where the <en-note> element has children.
		ExpectNoteText bool
		References     []TestReference
		ParentID       *TestReference
	}

	// ReferenceIDs is used to collect resource IDs that are either known ahead
	// of time or are generated after running the code.
	type ReferenceIDs struct{ Notebooks, Notes, Tags []string }

	testSNItem := func(t *testing.T, ind int, actual *sn.Item, expectedOutput []ExpectedTestValues) (ok bool) {
		t.Helper()
		ok = true
		if actual.Content.Title != expectedOutput[ind].Title {
			t.Errorf(
				"test %d; wrong title; got %q, expected %q",
				ind, actual.Content.Title, expectedOutput[ind].Title,
			)
			ok = false
		}
		if actual.ContentType.String() != expectedOutput[ind].ContentType {
			t.Errorf(
				"test %d; wrong content type; got %q, expected %q",
				ind, actual.ContentType, expectedOutput[ind].ContentType,
			)
			ok = false
		}
		if !uuidMatcher.MatchString(actual.UUID) {
			t.Errorf("test %d; invalid UUID: %q", ind, actual.UUID)
			ok = false
		}
		if actual.CreatedAt.IsZero() {
			t.Errorf("test %d; CreatedAt should not be zero", ind)
			ok = false
		}
		if actual.UpdatedAt.IsZero() {
			t.Errorf("test %d; UpdatedAt should not be zero", ind)
			ok = false
		}
		return
	}

	testSNNote := func(t *testing.T, ind int, actual *sn.Note, expectedOutput []ExpectedTestValues) (ok bool) {
		t.Helper()
		ok = true
		if expectedOutput[ind].ExpectNoteText && actual.Content.Text == "" {
			t.Errorf("test %d; expected note content not to be empty", ind)
			ok = false
		}
		val, hasKey := actual.Content.AppData["org.standardnotes.sn"]
		if !hasKey {
			t.Errorf(
				"test %d; expected to have value at key %q",
				ind, "org.standardnotes.sn",
			)
			ok = false
		}
		appData, ok := val.(*interactor.SNItemAppData)
		if !ok {
			t.Errorf(
				"test %d; appData should be a %T",
				ind, &interactor.SNItemAppData{},
			)
			ok = false
			return
		}
		if !appData.ClientUpdatedAt.Equal(actual.UpdatedAt) {
			t.Errorf(
				"test %d; app data incorrect; got %q, expected %q",
				ind, appData.ClientUpdatedAt, actual.UpdatedAt,
			)
			ok = false
		}
		return
	}

	testSNReferences := func(t *testing.T, ind int, tagRefs []sn.Reference, expected []TestReference, knownIDs *ReferenceIDs) (ok bool) {
		t.Helper()
		ok = true
		if len(tagRefs) != len(expected) {
			t.Errorf(
				"test %d; wrong number of tag references; got %d, expected %d",
				ind, len(tagRefs), len(expected),
			)
			ok = false
			return
		}
		for j, tag := range tagRefs {
			var expectedTagUUID string
			expTagRef := expected[j]

			switch tag.ContentType {
			case sn.ContentTypeTag:
				expectedTagUUID = knownIDs.Tags[expTagRef.Index]
			case sn.ContentTypeNotebook:
				expectedTagUUID = knownIDs.Notebooks[expTagRef.Index]
			case sn.ContentTypeNote:
				expectedTagUUID = knownIDs.Notes[expTagRef.Index]
			default:
				t.Errorf(
					"test [%d][%d]; unexpected type %q",
					ind, j, tag.ContentType,
				)
				ok = false
				continue
			}
			if tag.UUID != expectedTagUUID {
				t.Errorf(
					"test [%d][%d]; wrong tag UUID; got %q, expected %q",
					ind, j, tag.UUID, expectedTagUUID,
				)
				ok = false
			}
		}
		return
	}

	t.Run("EDAMToStandardNotes", func(t *testing.T) {
		out, err := interactor.ConvertEDAMToStandardNotes(
			context.TODO(),
			interactor.ConvertOptions{
				InputFilenames: struct{ Notebooks, Notes, Tags string }{
					Notebooks: _FixturesDir + "/" + _StubNotebooksFile,
					Notes:     _FixturesDir + "/" + _StubNotesFile,
					Tags:      _FixturesDir + "/" + _StubTagsFile,
				},
				OutputFilename: pathToTestDir + "/edam_to_standardnotes.json",
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		knownIDs := ReferenceIDs{
			Notes: []string{
				"8c44eeb1-7e50-4edb-95c4-12cf90d1017e", // Batman
				"7c56e278-c268-4003-b1cd-09853ad92b4a", // Atlanta
				"0f32f51b-f923-4dd0-b5c3-a7d6e8a8a40f", // Aladdin
				"9df1e45a-623d-4e1b-b370-2d2365499ed0", // Casino
				"9901a8a3-39a6-437c-9443-e60ef83e6394", // Baltimore
				"4a5704f8-0825-4926-8f9e-ca74b3c7da85", // Chicago
				"04630bf8-0800-408b-97d8-cebba0e8b864", // Despicable Me
				"c66bca64-4395-4675-ae86-9ef35cc0e5cf", // Enter The Dragon
				"e0322fce-4633-4d7d-8dff-79664844f03f", // Fargo
				"a2197031-1570-40e4-bc8f-0cb776057f6b", // Detroit
				"25480cfd-5785-4741-a6fd-a3e37aa9d43e", // Edmonton
				"879f5e58-60aa-496b-b764-bee8cfd664f6", // Fargo
				"1820018f-1d5e-4ae9-92c3-f0d72f45d25c", // Hello World
			},
			Tags: []string{
				"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564", // "foo"
				"f299e07c-b98e-4902-ac69-d0c7927e4870", // "free"
				"bb170464-72e0-4a22-85e0-b2c4f68272ea", // "bar"
				"53f1fdb5-4140-4ff4-8590-21e8cc2b4338", // "baker"
				"17585e5c-58cb-401c-b9db-d6f50c77993c", // "altered"
			},
			Notebooks: []string{
				"cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09", // "Cities"
				"932d7c12-bb87-4b41-895a-5d30fa178688", // "Movies"
				"6aac8caa-0682-4870-95fb-f384301704bc", // "<Inbox>"
				"ca62b4e6-5649-4512-ae30-2a8de03f80fe", // "Sample"
			},
		}

		expectedNotes := []ExpectedTestValues{
			{ContentType: "Note", Title: "Batman", ExpectNoteText: true, References: []TestReference{{"Tag", 0}, {"Tag", 3}, {"Notebook", 1}}},
			{ContentType: "Note", Title: "Atlanta", ExpectNoteText: true, References: []TestReference{{"Tag", 0}, {"Notebook", 0}}},
			{ContentType: "Note", Title: "Aladdin", References: []TestReference{{"Tag", 2}, {"Notebook", 1}}},
			{ContentType: "Note", Title: "Casino", References: []TestReference{{"Tag", 2}, {"Notebook", 1}}},
			{ContentType: "Note", Title: "Baltimore", References: []TestReference{{"Tag", 2}, {"Notebook", 0}}},
			{ContentType: "Note", Title: "Chicago", References: []TestReference{{"Tag", 1}, {"Tag", 3}, {"Notebook", 0}}},
			{ContentType: "Note", Title: "Despicable Me", References: []TestReference{{"Tag", 0}, {"Notebook", 1}}},
			{ContentType: "Note", Title: "Enter The Dragon", References: []TestReference{{"Tag", 2}, {"Notebook", 1}}},
			{ContentType: "Note", Title: "Fargo", References: []TestReference{{"Tag", 0}, {"Notebook", 1}}},
			{ContentType: "Note", Title: "Detroit", References: []TestReference{{"Tag", 2}, {"Notebook", 0}}},
			{ContentType: "Note", Title: "Edmonton", References: []TestReference{{"Tag", 0}, {"Notebook", 0}}},
			{ContentType: "Note", Title: "Fargo", References: []TestReference{{"Tag", 2}, {"Notebook", 0}}},
			{ContentType: "Note", Title: "Hello World", ExpectNoteText: true, References: []TestReference{{"Tag", 0}, {"Tag", 3}, {"Notebook", 2}}},
		}
		expectedTags := []ExpectedTestValues{
			{ContentType: "Tag", Title: "foo", References: []TestReference{{"Note", 0}, {"Note", 1}, {"Note", 6}, {"Note", 8}, {"Note", 10}, {"Note", 12}}},
			{ContentType: "Tag", Title: "free", References: []TestReference{{"Note", 5}}, ParentID: &TestReference{"Tag", 0}},
			{ContentType: "Tag", Title: "bar", References: []TestReference{{"Note", 2}, {"Note", 3}, {"Note", 4}, {"Note", 7}, {"Note", 9}, {"Note", 11}}},
			{ContentType: "Tag", Title: "baker", References: []TestReference{{"Note", 0}, {"Note", 5}, {"Note", 12}}, ParentID: &TestReference{"Tag", 2}},
			{ContentType: "Tag", Title: "altered", References: []TestReference{}},
		}
		expectedNotebooks := []ExpectedTestValues{
			{ContentType: "Notebook", Title: "Cities", References: []TestReference{{"Note", 1}, {"Note", 4}, {"Note", 5}, {"Note", 9}, {"Note", 10}, {"Note", 11}}},
			{ContentType: "Notebook", Title: "Movies", References: []TestReference{{"Note", 0}, {"Note", 2}, {"Note", 3}, {"Note", 6}, {"Note", 7}, {"Note", 8}}},
			{ContentType: "Notebook", Title: "<Inbox>", References: []TestReference{{"Hello World", 12}}},
			{ContentType: "Notebook", Title: "Samples", References: []TestReference{}},
		}
		expectedOutput := make([]ExpectedTestValues, 0)
		expectedOutput = append(expectedOutput, expectedNotes...)
		expectedOutput = append(expectedOutput, expectedTags...)
		expectedOutput = append(expectedOutput, expectedNotebooks...)

		actualItems := out.Items
		if len(actualItems) != len(expectedOutput) {
			t.Fatalf(
				"wrong output length; got %d, expected  %d",
				len(actualItems), len(expectedOutput),
			)
		}

		for ind, member := range actualItems {
			switch actual := member.(type) {
			case *sn.Note:
				testSNItem(t, ind, &actual.Item, expectedOutput)
				testSNNote(t, ind, actual, expectedOutput)
				testSNReferences(t, ind, actual.Content.References, expectedOutput[ind].References, &knownIDs)
			case *sn.Tag:
				testSNItem(t, ind, &actual.Item, expectedOutput)
				testSNReferences(t, ind, actual.Content.References, expectedOutput[ind].References, &knownIDs)
			default:
				t.Fatalf("test %d; unexpected type %T", ind, actual)
			}
		}

		numExpectedNotes := len(expectedNotes)
		numExpectedTags := len(expectedTags)

		actualTags := actualItems[numExpectedNotes : numExpectedNotes+numExpectedTags]
		for ind, member := range actualTags {
			tag, ok := member.(*sn.Tag)
			if !ok {
				t.Errorf("test %d; expected a %T", ind, &sn.Tag{})
				continue
			}

			val, ok := tag.Content.AppData["evernote.com"]
			if !ok {
				t.Errorf("test %d; expected appData to have key %q", ind, "evernote")
				continue
			}

			appData, ok := val.(*interactor.SNItemAppData)
			if !ok {
				t.Errorf(
					"test %d; appData should be a %T",
					ind, &interactor.SNItemAppData{},
				)
				continue
			}

			if expectedTags[ind].ParentID == nil {
				continue
			}
			expected := knownIDs.Tags[expectedTags[ind].ParentID.Index]
			if appData.ParentID != expected {
				t.Errorf(
					"test %d; app data incorrect; got %q, expected %q",
					ind, appData.ParentID, expected,
				)
			}
		}

		// test notebooks separately because there are some workarounds for
		// StandardNote's lack of Notebooks.
		actualNotebooks := actualItems[numExpectedNotes+numExpectedTags:]
		for ind, member := range actualNotebooks {
			tag, ok := member.(*sn.Tag)
			if !ok {
				t.Errorf("test %d; expected a %T", ind, &sn.Tag{})
				continue
			}
			val, ok := tag.Content.AppData["evernote.com"]
			if !ok {
				t.Errorf("test %d; expected appData to have key %q", ind, "evernote")
			}

			appData, ok := val.(*interactor.SNItemAppData)
			if !ok {
				t.Errorf(
					"test %d; appData should be a %T",
					ind, &interactor.SNItemAppData{},
				)
				continue
			}
			if appData.OriginalContentType != "Notebook" {
				t.Errorf(
					"test %d; app data incorrect; got %q, expected %q",
					ind, appData.OriginalContentType, "Notebook",
				)
			}
		}
	})

	t.Run("ENEXToStandardNotes", func(t *testing.T) {
		out, err := interactor.ConvertENEXToStandardNotes(
			context.TODO(),
			interactor.ConvertOptions{
				InputFilename:  _FixturesDir + "/" + _StubENEXFile,
				OutputFilename: pathToTestDir + "/enex_to_standardnotes.json",
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// collect generated tag info to use in tests later.
		knownIDs := ReferenceIDs{
			Notes: make([]string, 0),
			Tags:  make([]string, 0),
		}
		actualItems := out.Items
		for ind, member := range actualItems {
			switch actual := member.(type) {
			case *sn.Note:
				knownIDs.Notes = append(knownIDs.Notes, actual.UUID)
			case *sn.Tag:
				knownIDs.Tags = append(knownIDs.Tags, actual.UUID)
			default:
				t.Fatalf("%d, unexpected type %T", ind, actual)
			}
		}

		baseExpectedNotes := []ExpectedTestValues{
			{ContentType: "Note", Title: "Batman", ExpectNoteText: true, References: []TestReference{{"Tag", 0}, {"Tag", 1}}},
			{ContentType: "Note", Title: "Atlanta", ExpectNoteText: true, References: []TestReference{{"Tag", 0}}},
			{ContentType: "Note", Title: "Hello World", ExpectNoteText: true, References: []TestReference{{"Tag", 0}, {"Tag", 1}}},
			{ContentType: "Note", Title: "Chicago", References: []TestReference{{"Tag", 1}, {"Tag", 2}}},
			{ContentType: "Note", Title: "Edmonton", References: []TestReference{{"Tag", 0}}},
			{ContentType: "Note", Title: "Baltimore", References: []TestReference{{"Tag", 3}}},
			{ContentType: "Note", Title: "Fargo", References: []TestReference{{"Tag", 3}}},
			{ContentType: "Note", Title: "Despicable Me", References: []TestReference{{"Tag", 0}}},
			{ContentType: "Note", Title: "Fargo", References: []TestReference{{"Tag", 0}}},
			{ContentType: "Note", Title: "Enter The Dragon", References: []TestReference{{"Tag", 3}}},
			{ContentType: "Note", Title: "Detroit", References: []TestReference{{"Tag", 3}}},
			{ContentType: "Note", Title: "Casino", References: []TestReference{{"Tag", 3}}},
			{ContentType: "Note", Title: "Aladdin", References: []TestReference{{"Tag", 3}}},
		}

		baseExpectedTags := []ExpectedTestValues{
			{ContentType: "Tag", Title: "foo", References: []TestReference{{"Note", 0}, {"Note", 1}, {"Note", 2}, {"Note", 4}, {"Note", 7}, {"Note", 8}}},
			{ContentType: "Tag", Title: "baker", References: []TestReference{{"Note", 0}, {"Note", 2}, {"Note", 3}}},
			{ContentType: "Tag", Title: "free", References: []TestReference{{"Note", 3}}},
			{ContentType: "Tag", Title: "bar", References: []TestReference{{"Note", 5}, {"Note", 6}, {"Note", 9}, {"Note", 10}, {"Note", 11}, {"Note", 12}}},
		}
		expectedOutput := make([]ExpectedTestValues, 0)
		expectedOutput = append(expectedOutput, baseExpectedNotes...)
		expectedOutput = append(expectedOutput, baseExpectedTags...)

		if len(actualItems) != len(expectedOutput) {
			t.Fatalf(
				"wrong output length; got %d, expected  %d",
				len(actualItems), len(expectedOutput),
			)
		}

		for ind, member := range actualItems {
			switch actual := member.(type) {
			case *sn.Note:
				testSNItem(t, ind, &actual.Item, expectedOutput)
				testSNNote(t, ind, actual, expectedOutput)
				testSNReferences(t, ind, actual.Content.References, expectedOutput[ind].References, &knownIDs)
			case *sn.Tag:
				testSNItem(t, ind, &actual.Item, expectedOutput)
				testSNReferences(t, ind, actual.Content.References, expectedOutput[ind].References, &knownIDs)
			default:
				t.Errorf("test %d; unexpected type %T", ind, actual)
			}
		}
	})
}

// uuidMatcher helps us make sure we're at least trying to make a UUID. Pattern
// lifted from: https://stackoverflow.com/a/13653180.
var uuidMatcher = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestReconciliation(t *testing.T) {
	pathToTestDir := _BaseTestOutputDir + "/" + t.Name()
	if err := os.MkdirAll(pathToTestDir, 0755); err != nil {
		t.Fatal(err)
	}
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
