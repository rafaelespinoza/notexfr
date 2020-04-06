package entity_test

import (
	"testing"

	"github.com/rafaelespinoza/snbackfill/internal/entity"
)

func TestInterfaceImplementations(t *testing.T) {
	implementations := []interface{}{
		new(entity.ServiceID),
	}
	for i, val := range implementations {
		if _, ok := val.(entity.Resource); !ok {
			t.Errorf(
				"test %d; expected value of type %T to implement entity.Resource",
				i, val,
			)
		}
	}
}
