package edam_test

import (
	"context"
	"os"
	"testing"
	"time"

	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/entity"
	"github.com/rafaelespinoza/snbackfill/internal/repo/edam"
)

const (
	// _FixturesDir should be relative to this file's directory.
	_FixturesDir       = "../../../internal/fixtures"
	_StubNotebooksFile = "edam_notebooks.json"
	_StubNotesFile     = "edam_notes.json"
	_StubTagsFile      = "edam_tags.json"
)

func TestInterfaceImplementations(t *testing.T) {
	t.Run("collections", func(t *testing.T) {
		implementations := []interface{}{
			new(edam.Notebooks),
			new(edam.Notes),
			new(edam.Tags),
		}
		for i, val := range implementations {
			if _, ok := val.(lib.LocalRemoteRepo); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement lib.LocalRemoteRepo",
					i, val,
				)
			}
		}
	})

	t.Run("members", func(t *testing.T) {
		implementations := []interface{}{
			new(edam.Notebook),
			new(edam.Note),
			new(edam.Tag),
		}
		for i, val := range implementations {
			var ok bool
			if _, ok = val.(lib.LinkID); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement lib.LinkID",
					i, val,
				)
			}
		}
	})
}

func TestNotebooks(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		var (
			repo            lib.LocalRemoteRepo
			err             error
			actualResources []lib.LinkID
			resource        *edam.Notebook
			ok              bool
		)
		repo, _ = edam.NewNotebooksRepo()
		if actualResources, err = readLocalFile(repo, _FixturesDir+"/"+_StubNotebooksFile); err != nil {
			t.Fatal(err)
		}
		expectedResources := []*entity.Notebook{
			{
				ID:        "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				Name:      "Cities",
				Stack:     "",
				CreatedAt: time.Date(2020, 3, 7, 2, 14, 49, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 23, 35, 0, time.UTC),
			},
			{
				ID:        "932d7c12-bb87-4b41-895a-5d30fa178688",
				Name:      "Movies",
				Stack:     "",
				CreatedAt: time.Date(2020, 3, 7, 2, 14, 58, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 7, 20, 23, 55, 0, time.UTC),
			},
			{
				ID:        "6aac8caa-0682-4870-95fb-f384301704bc",
				Name:      "<Inbox>",
				Stack:     "SampleStack Stack",
				CreatedAt: time.Date(2020, 3, 7, 2, 13, 22, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 8, 22, 17, 53, 0, time.UTC),
			},
			{
				ID:        "ca62b4e6-5649-4512-ae30-2a8de03f80fe",
				Name:      "Samples",
				Stack:     "SampleStack Stack",
				CreatedAt: time.Date(2020, 3, 8, 22, 16, 20, 0, time.UTC),
				UpdatedAt: time.Date(2020, 3, 8, 22, 44, 6, 0, time.UTC),
			},
		}
		if len(actualResources) != len(expectedResources) {
			t.Fatalf(
				"wrong length; got %d expected %d",
				len(actualResources), len(expectedResources),
			)
		}

		for i, res := range actualResources {
			if resource, ok = res.(*edam.Notebook); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, resource, &entity.Notebook{},
				)
			}
			if resource.ID != expectedResources[i].ID {
				t.Errorf(
					"test %d; wrong ID; got %q, expected %q",
					i, resource.ID, expectedResources[i].ID,
				)
			}
			if resource.Name != expectedResources[i].Name {
				t.Errorf(
					"test %d; wrong Name; got %q, expected %q",
					i, resource.Name, expectedResources[i].Name,
				)
			}
			if resource.Stack != expectedResources[i].Stack {
				t.Errorf(
					"test %d; wrong Stack; got %q, expected %q",
					i, resource.Stack, expectedResources[i].Stack,
				)
			}
			if !resource.CreatedAt.Equal(expectedResources[i].CreatedAt) {
				t.Errorf(
					"test %d; wrong CreatedAt; got %q, expected %q",
					i, resource.CreatedAt, expectedResources[i].CreatedAt,
				)
			}
			if !resource.UpdatedAt.Equal(expectedResources[i].UpdatedAt) {
				t.Errorf(
					"test %d; wrong UpdatedAt; got %q, expected %q",
					i, resource.UpdatedAt, expectedResources[i].UpdatedAt,
				)
			}
		}
	})
}

func TestNotes(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		var (
			repo            lib.LocalRemoteRepo
			err             error
			actualResources []lib.LinkID
			resource        *edam.Note
			ok              bool
		)
		repo, _ = edam.NewNotesRepo(nil)
		if actualResources, err = readLocalFile(repo, _FixturesDir+"/"+_StubNotesFile); err != nil {
			t.Fatal(err)
		}
		expectedResources := []*entity.Note{
			{
				Attributes: &entity.Attributes{
					Source: "desktop.mac",
				},
				CreatedAt:  time.Date(2020, 3, 8, 22, 18, 21, 0, time.UTC),
				ID:         "1820018f-1d5e-4ae9-92c3-f0d72f45d25c",
				NotebookID: "6aac8caa-0682-4870-95fb-f384301704bc",
				TagIDs: []string{
					"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
					"53f1fdb5-4140-4ff4-8590-21e8cc2b4338",
				},
				Tags:      nil,
				Title:     "Hello World",
				UpdatedAt: time.Date(2020, 3, 8, 22, 18, 36, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 33, 36, 0, time.UTC),
				ID:         "879f5e58-60aa-496b-b764-bee8cfd664f6",
				NotebookID: "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				TagIDs: []string{
					"bb170464-72e0-4a22-85e0-b2c4f68272ea",
				},
				Tags:      nil,
				Title:     "Fargo",
				UpdatedAt: time.Date(2020, 3, 7, 20, 34, 48, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 22, 13, 0, time.UTC),
				ID:         "7c56e278-c268-4003-b1cd-09853ad92b4a",
				NotebookID: "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				TagIDs: []string{
					"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
				},
				Tags:      nil,
				Title:     "Atlanta",
				UpdatedAt: time.Date(2020, 3, 7, 20, 33, 33, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 33, 13, 0, time.UTC),
				ID:         "25480cfd-5785-4741-a6fd-a3e37aa9d43e",
				NotebookID: "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				TagIDs: []string{
					"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
				},
				Tags:      nil,
				Title:     "Edmonton",
				UpdatedAt: time.Date(2020, 3, 7, 20, 33, 21, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 32, 37, 0, time.UTC),
				ID:         "a2197031-1570-40e4-bc8f-0cb776057f6b",
				NotebookID: "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				TagIDs: []string{
					"bb170464-72e0-4a22-85e0-b2c4f68272ea",
				},
				Tags:      nil,
				Title:     "Detroit",
				UpdatedAt: time.Date(2020, 3, 7, 20, 32, 43, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 30, 42, 0, time.UTC),
				ID:         "e0322fce-4633-4d7d-8dff-79664844f03f",
				NotebookID: "932d7c12-bb87-4b41-895a-5d30fa178688",
				TagIDs: []string{
					"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
				},
				Tags:      nil,
				Title:     "Fargo",
				UpdatedAt: time.Date(2020, 3, 7, 20, 30, 50, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 30, 5, 0, time.UTC),
				ID:         "c66bca64-4395-4675-ae86-9ef35cc0e5cf",
				NotebookID: "932d7c12-bb87-4b41-895a-5d30fa178688",
				TagIDs: []string{
					"bb170464-72e0-4a22-85e0-b2c4f68272ea",
				},
				Tags:      nil,
				Title:     "Enter The Dragon",
				UpdatedAt: time.Date(2020, 3, 7, 20, 30, 9, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 29, 39, 0, time.UTC),
				ID:         "04630bf8-0800-408b-97d8-cebba0e8b864",
				NotebookID: "932d7c12-bb87-4b41-895a-5d30fa178688",
				TagIDs: []string{
					"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
				},
				Tags:      nil,
				Title:     "Despicable Me",
				UpdatedAt: time.Date(2020, 3, 7, 20, 29, 44, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 26, 3, 0, time.UTC),
				ID:         "4a5704f8-0825-4926-8f9e-ca74b3c7da85",
				NotebookID: "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				TagIDs: []string{
					"f299e07c-b98e-4902-ac69-d0c7927e4870",
					"53f1fdb5-4140-4ff4-8590-21e8cc2b4338",
				},
				Tags:      nil,
				Title:     "Chicago",
				UpdatedAt: time.Date(2020, 3, 7, 20, 26, 9, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 21, 56, 0, time.UTC),
				ID:         "8c44eeb1-7e50-4edb-95c4-12cf90d1017e",
				NotebookID: "932d7c12-bb87-4b41-895a-5d30fa178688",
				TagIDs: []string{
					"ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
					"53f1fdb5-4140-4ff4-8590-21e8cc2b4338",
				},
				Tags:      nil,
				Title:     "Batman",
				UpdatedAt: time.Date(2020, 3, 7, 20, 25, 54, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 25, 24, 0, time.UTC),
				ID:         "9901a8a3-39a6-437c-9443-e60ef83e6394",
				NotebookID: "cdb30948-fd4b-4f0f-88e8-68f0ed9e5a09",
				TagIDs: []string{
					"bb170464-72e0-4a22-85e0-b2c4f68272ea",
				},
				Tags:      nil,
				Title:     "Baltimore",
				UpdatedAt: time.Date(2020, 3, 7, 20, 25, 29, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 25, 10, 0, time.UTC),
				ID:         "9df1e45a-623d-4e1b-b370-2d2365499ed0",
				NotebookID: "932d7c12-bb87-4b41-895a-5d30fa178688",
				TagIDs: []string{
					"bb170464-72e0-4a22-85e0-b2c4f68272ea",
				},
				Tags:      nil,
				Title:     "Casino",
				UpdatedAt: time.Date(2020, 3, 7, 20, 25, 16, 0, time.UTC),
			},
			{
				Attributes: &entity.Attributes{},
				CreatedAt:  time.Date(2020, 3, 7, 20, 24, 03, 0, time.UTC),
				ID:         "0f32f51b-f923-4dd0-b5c3-a7d6e8a8a40f",
				NotebookID: "932d7c12-bb87-4b41-895a-5d30fa178688",
				TagIDs: []string{
					"bb170464-72e0-4a22-85e0-b2c4f68272ea",
				},
				Tags:      nil,
				Title:     "Aladdin",
				UpdatedAt: time.Date(2020, 3, 7, 20, 24, 47, 0, time.UTC),
			},
		}
		if len(actualResources) != len(expectedResources) {
			t.Fatalf(
				"wrong length; got %d expected %d",
				len(actualResources), len(expectedResources),
			)
		}
		for i, res := range actualResources {
			if resource, ok = res.(*edam.Note); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, resource, &entity.Note{},
				)
			}
			if resource.ID != expectedResources[i].ID {
				t.Errorf(
					"test %d; wrong ID; got %q, expected %q",
					i, resource.ID, expectedResources[i].ID,
				)
			}
			if resource.Title != expectedResources[i].Title {
				t.Errorf(
					"test %d; wrong Title; got %q, expected %q",
					i, resource.Title, expectedResources[i].Title,
				)
			}
			if resource.NotebookID != expectedResources[i].NotebookID {
				t.Errorf(
					"test %d; wrong NotebookID; got %q, expected %q",
					i, resource.NotebookID, expectedResources[i].NotebookID,
				)
			}
			if !resource.CreatedAt.Equal(expectedResources[i].CreatedAt) {
				t.Errorf(
					"test %d; wrong CreatedAt; got %q, expected %q",
					i, resource.CreatedAt, expectedResources[i].CreatedAt,
				)
			}
			if !resource.UpdatedAt.Equal(expectedResources[i].UpdatedAt) {
				t.Errorf(
					"test %d; wrong UpdatedAt; got %q, expected %q",
					i, resource.UpdatedAt, expectedResources[i].UpdatedAt,
				)
			}
			if len(resource.TagIDs) != len(expectedResources[i].TagIDs) {
				t.Errorf(
					"test %d; wrong length for TagIDs; got %d, expected %d",
					i, len(resource.TagIDs), len(expectedResources[i].TagIDs),
				)
			}
			for j, tagID := range resource.TagIDs {
				if tagID != expectedResources[i].TagIDs[j] {
					t.Errorf(
						"test [%d][%d] wrong tagID; got %q, expected %q",
						i, j, tagID, expectedResources[i].TagIDs[j],
					)
				}
			}
			if len(resource.Tags) != len(expectedResources[i].Tags) {
				t.Errorf(
					"test %d; wrong length for Tags; got %d, expected %d",
					i, len(resource.Tags), len(expectedResources[i].Tags),
				)
			}
			for j, tag := range resource.Tags {
				if tag != expectedResources[i].Tags[j] {
					t.Errorf(
						"test [%d][%d] wrong tag; got %q, expected %q",
						i, j, tag, expectedResources[i].Tags[j],
					)
				}
			}
			// it's a pointer type, not sure this makes sense...
			if *resource.Attributes != *expectedResources[i].Attributes {
				t.Errorf(
					"test %d; wrong Attributes; got %q, expected %q",
					i, *resource.Attributes, *expectedResources[i].Attributes,
				)
			}
		}
	})
}

func TestTags(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		var (
			repo            lib.LocalRemoteRepo
			err             error
			actualResources []lib.LinkID
			resource        *edam.Tag
			ok              bool
		)
		repo, _ = edam.NewTagsRepo()
		if actualResources, err = readLocalFile(repo, _FixturesDir+"/"+_StubTagsFile); err != nil {
			t.Fatal(err)
		}
		expectedResources := []*entity.Tag{
			{
				ID:       "ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
				Name:     "foo",
				ParentID: "",
			},
			{
				ID:       "f299e07c-b98e-4902-ac69-d0c7927e4870",
				Name:     "free",
				ParentID: "ed18a1cf-e1f7-4d51-8d9b-f7201e60f564",
			},
			{
				ID:       "bb170464-72e0-4a22-85e0-b2c4f68272ea",
				Name:     "bar",
				ParentID: "",
			},
			{
				ID:       "53f1fdb5-4140-4ff4-8590-21e8cc2b4338",
				Name:     "baker",
				ParentID: "bb170464-72e0-4a22-85e0-b2c4f68272ea",
			},
			{
				ID:       "17585e5c-58cb-401c-b9db-d6f50c77993c",
				Name:     "altered",
				ParentID: "",
			},
		}
		if len(actualResources) != len(expectedResources) {
			t.Fatalf(
				"wrong length; got %d expected %d",
				len(actualResources), len(expectedResources),
			)
		}

		for i, res := range actualResources {
			if resource, ok = res.(*edam.Tag); !ok {
				t.Fatalf(
					"test %d; wrong type; got %T, expected %T",
					i, resource, &edam.Tag{},
				)
			}
			if resource.ID != expectedResources[i].ID {
				t.Errorf(
					"test %d; wrong ID; got %q, expected %q",
					i, resource.ID, expectedResources[i].ID,
				)
			}
			if resource.Name != expectedResources[i].Name {
				t.Errorf(
					"test %d; wrong Name; got %q, expected %q",
					i, resource.Name, expectedResources[i].Name,
				)
			}
			if resource.ParentID != expectedResources[i].ParentID {
				t.Errorf(
					"test %d; wrong ParentID; got %q, expected %q",
					i, resource.ParentID, expectedResources[i].ParentID,
				)
			}
		}
	})
}

func readLocalFile(repository lib.RepoLocal, filename string) (out []lib.LinkID, err error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	out, err = repository.ReadLocal(context.TODO(), file)
	return
}
