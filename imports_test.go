// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type importsSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&importsSuite{})

var importsTests = []struct {
	pkgName string
	prefix  string
	expect  []string
}{{
	pkgName: "github.com/juju/testing",
	prefix:  "github.com/juju/testing/",
	expect:  []string{"checkers"},
}, {
	pkgName: "github.com/juju/testing",
	prefix:  "github.com/juju/utils/v2/",
	expect:  []string{},
}, {
	pkgName: "github.com/juju/testing",
	prefix:  "arble.com/",
	expect:  nil,
}}

func (s *importsSuite) TestImports(c *gc.C) {
	for i, test := range importsTests {
		c.Logf("test %d: %s %s", i, test.pkgName, test.prefix)
		imports, err := testing.FindImports(test.pkgName, test.prefix)
		c.Assert(err, gc.IsNil)
		c.Assert(imports, jc.DeepEquals, test.expect)
	}
}
