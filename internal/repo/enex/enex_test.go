package enex_test

import (
	"testing"

	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/repo/enex"
)

func TestInterfaceImplementations(t *testing.T) {
	t.Run("collections", func(t *testing.T) {
		implementations := []interface{}{
			new(enex.File),
		}
		for i, val := range implementations {
			if _, ok := val.(lib.RepoLocal); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement lib.RepoLocal",
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
			if _, ok := val.(lib.LinkID); !ok {
				t.Errorf(
					"test %d; expected value of type %T to implement lib.LinkID",
					i, val,
				)
			}
		}
	})
}
