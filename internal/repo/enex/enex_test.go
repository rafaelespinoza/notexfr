package enex_test

import (
	"testing"

	"github.com/rafaelespinoza/snbackfill/internal/entity"
	"github.com/rafaelespinoza/snbackfill/internal/repo/enex"
)

func TestInterfaceImplementations(t *testing.T) {
	t.Run("collections", func(t *testing.T) {
		implementations := []interface{}{
			new(enex.File),
		}
		for i, val := range implementations {
			if _, ok := val.(entity.RepoLocal); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement entity.RepoLocal",
					i, val,
				)
			}
		}
	})

	t.Run("members", func(t *testing.T) {
		implementations := []interface{}{
			new(enex.Note),
		}
		for i, val := range implementations {
			if _, ok := val.(entity.LinkID); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement entity.LinkID",
					i, val,
				)
			}
		}
	})
}
