package service_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"vervet-underground/internal/service"
)

func TestRegistry_Load(t *testing.T) {
	c := qt.New(t)

	ldrReturns := [][]string{
		{"a", "b", "c"},
		{"a"},
	}
	ldr := func(context.Context) ([]string, error) {
		var ret []string
		ret, ldrReturns = ldrReturns[0], ldrReturns[1:]
		return ret, nil
	}

	reg := service.NewRegistry(ldr)
	c.Assert(len(reg.Services), qt.Equals, 0)

	// 3 services loaded
	c.Assert(reg.Load(), qt.IsNil)
	c.Assert(len(reg.Services), qt.Equals, 3)

	// 1 service loaded
	c.Assert(reg.Load(), qt.IsNil)
	c.Assert(len(reg.Services), qt.Equals, 1)
}
