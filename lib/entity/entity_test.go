package entity_test

import (
	"testing"

	"github.com/rafaelespinoza/snbackfill/lib"
	"github.com/rafaelespinoza/snbackfill/lib/entity"
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
