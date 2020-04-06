package sn_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rafaelespinoza/snbackfill/internal/entity"
	"github.com/rafaelespinoza/snbackfill/internal/repo/sn"
)

const (
	// _FixturesDir should be relative to this file's directory.
	_FixturesDir    = "../../../internal/fixtures"
	_StubENtoSNFile = "evernote-to-sn.txt"
)

func TestInterfaceImplementations(t *testing.T) {
	var implementations []interface{}

	implementations = []interface{}{
		new(sn.Note),
		new(sn.Tag),
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
}

func TestReadConversionFile(t *testing.T) {
	const pathToFile = _FixturesDir + "/" + _StubENtoSNFile
	var (
		notes []entity.LinkID
		tags  []entity.LinkID
		err   error
		note  *sn.Note
		tag   *sn.Tag
		ok    bool
	)
	if notes, tags, err = sn.ReadConversionFile(pathToFile); err != nil {
		t.Fatal(err)
	}

	// expectedTestValues is a set of expected values for one test case.
	type expectedTestValues struct {
		UUID      string
		CreatedAt time.Time
		UpdatedAt time.Time
		Title     string
	}

	expectedTags := []expectedTestValues{
		{
			UUID:  "845e98ed-4515-473e-836b-ada5b5cb8d01",
			Title: "foo",
		},
		{
			UUID:  "4cbfa0b5-655c-447c-9961-6a5294b6b041",
			Title: "baker",
		},
		{
			UUID:  "e5daa664-db99-4cbf-afe5-2c0f043bac8c",
			Title: "free",
		},
		{
			UUID:  "30cf9510-845d-4ea8-b673-51104a3e0bc2",
			Title: "bar",
		},
		{
			UUID:  "90e045a2-46ea-44ff-808b-648274926c7f",
			Title: "evernote",
		},
	}

	for i, link := range tags {
		if tag, ok = link.(*sn.Tag); !ok {
			t.Fatalf(
				"test %d; wrong type; got %T, expected %T",
				i, link, &sn.Tag{},
			)
		}
		if tag.ServiceID == nil {
			t.Error("did not expect ServiceID to be nil")
			continue
		}
		if tag.GetID() != expectedTags[i].UUID {
			t.Errorf(
				"test %d; wrong ID; got %q, expected %q",
				i, tag.GetID(), expectedTags[i].UUID,
			)
		}
		if tag.Content.Title != expectedTags[i].Title {
			t.Errorf(
				"test %d; wrong Title; got %q, expected %q",
				i, tag.Content.Title, expectedTags[i].Title,
			)
		}
		// We don't care about timestamp values here because the EDAM API does
		// not have it for Tag and the value in the StandardNote conversion
		// files is the time it was converted, which is not relevant.
	}

	expectedNotes := []expectedTestValues{
		{
			UUID:      "8e053669-d1cc-4b69-a7fd-4433fc48feb7",
			CreatedAt: time.Date(2020, 3, 7, 20, 21, 56, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 25, 54, 0, time.UTC),
			Title:     "Batman",
		},
		{
			UUID:      "8154cd4e-dd06-4386-afe4-ec09f847b708",
			CreatedAt: time.Date(2020, 3, 7, 20, 22, 13, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 33, 33, 0, time.UTC),
			Title:     "Atlanta",
		},
		{
			UUID:      "228e48b8-1f46-4c79-a429-a093ed21656c",
			CreatedAt: time.Date(2020, 3, 8, 22, 18, 21, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 8, 22, 18, 36, 0, time.UTC),
			Title:     "Hello World",
		},
		{
			UUID:      "148dbae4-14b7-420e-8cb0-448d1be90ec6",
			CreatedAt: time.Date(2020, 3, 7, 20, 26, 3, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 26, 9, 0, time.UTC),
			Title:     "Chicago",
		},
		{
			UUID:      "7d43e417-d6f2-4a7e-9927-60d914c75a45",
			CreatedAt: time.Date(2020, 3, 7, 20, 33, 13, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 33, 21, 0, time.UTC),
			Title:     "Edmonton",
		},
		{
			UUID:      "16245526-2f33-4024-ab55-f6c17a61c053",
			CreatedAt: time.Date(2020, 3, 7, 20, 25, 24, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 25, 29, 0, time.UTC),
			Title:     "Baltimore",
		},
		{
			UUID:      "9c78d572-bf80-4fd5-98ce-5b21624100fe",
			CreatedAt: time.Date(2020, 3, 7, 20, 33, 36, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 34, 48, 0, time.UTC),
			Title:     "Fargo",
		},
		{
			UUID:      "27822e0b-7e2e-4ef3-ab8b-164f7932abfb",
			CreatedAt: time.Date(2020, 3, 7, 20, 29, 39, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 29, 44, 0, time.UTC),
			Title:     "Despicable Me",
		},
		{
			UUID:      "3d378ee3-774f-49cc-838d-4fec925c593a",
			CreatedAt: time.Date(2020, 3, 7, 20, 30, 42, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 30, 50, 0, time.UTC),
			Title:     "Fargo",
		},
		{
			UUID:      "581a0e66-2cbf-4976-b5e2-a5c6705ca0af",
			CreatedAt: time.Date(2020, 3, 7, 20, 30, 5, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 30, 9, 0, time.UTC),
			Title:     "Enter The Dragon",
		},
		{
			UUID:      "59f60f6f-fa97-4eaa-80a3-0006e345d381",
			CreatedAt: time.Date(2020, 3, 7, 20, 32, 37, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 32, 43, 0, time.UTC),
			Title:     "Detroit",
		},
		{
			UUID:      "0f4fe851-b6ff-455a-8d10-cc2e441f7deb",
			CreatedAt: time.Date(2020, 3, 7, 20, 25, 10, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 25, 16, 0, time.UTC),
			Title:     "Casino",
		},
		{
			UUID:      "63f6d85b-1ac8-4ccf-8ae7-58007fcd033d",
			CreatedAt: time.Date(2020, 3, 7, 20, 24, 3, 0, time.UTC),
			UpdatedAt: time.Date(2020, 3, 7, 20, 24, 47, 0, time.UTC),
			Title:     "Aladdin",
		},
	}
	for i, link := range notes {
		if note, ok = link.(*sn.Note); !ok {
			t.Fatalf(
				"test %d; wrong type; got %T, expected %T",
				i, tag, &sn.Tag{},
			)
		}
		if note.ServiceID == nil {
			t.Error("did not expect ServiceID to be nil")
			continue
		}
		if note.GetID() != expectedNotes[i].UUID {
			t.Errorf(
				"test %d; wrong ID; got %q, expected %q",
				i, note.GetID(), expectedNotes[i].UUID,
			)
		}
		if !note.CreatedAt.Equal(expectedNotes[i].CreatedAt) {
			t.Errorf(
				"test %d; wrong CreatedAt; got %q, expected %q",
				i, note.CreatedAt, expectedNotes[i].CreatedAt,
			)
		}
		if !note.UpdatedAt.Equal(expectedNotes[i].UpdatedAt) {
			t.Errorf(
				"test %d; wrong UpdatedAt; got %q, expected %q",
				i, note.UpdatedAt, expectedNotes[i].UpdatedAt,
			)
		}
		if note.Content.Title != expectedNotes[i].Title {
			t.Errorf(
				"test %d; wrong Title; got %q, expected %q",
				i, note.Content.Title, expectedNotes[i].Title,
			)
		}
	}
	// A smoke test (should not panic). Also for manual overview.
	mustPrettyPrintJSON(t, notes)
}

func mustPrettyPrintJSON(t *testing.T, in interface{}) {
	out, err := json.MarshalIndent(in, "  ", "    ")
	if err != nil {
		panic(err)
	}
	t.Logf("%+v\n", string(out))
}
