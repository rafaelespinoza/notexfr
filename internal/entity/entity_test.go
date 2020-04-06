package entity_test

import (
	"testing"

	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/entity"
)

func TestInterfaceImplementations(t *testing.T) {
	implementations := []interface{}{
		new(entity.ServiceID),
	}
	for i, val := range implementations {
		if _, ok := val.(lib.Resource); !ok {
			t.Errorf(
				"test %d; expected value of type %T to implement lib.Resource",
				i, val,
			)
		}
	}
}
