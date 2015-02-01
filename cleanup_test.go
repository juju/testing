// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"os"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
)

type cleanupSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&cleanupSuite{})

func (s *cleanupSuite) TestTearDownSuiteEmpty(c *gc.C) {
	suite := testing.CleanupSuite{}
	suite.TearDownSuite(c)
	suite.SetUpSuite(c)
}

func (s *cleanupSuite) TestTearDownTestEmpty(c *gc.C) {
	suite := testing.CleanupSuite{}
	suite.TearDownTest(c)
	suite.SetUpTest(c)
}

func (s *cleanupSuite) TestTearDownTestWithPatch(c *gc.C) {
	expSuiteVal := 42
	expTestVal := 84
	dest := 0
	suite := testing.CleanupSuite{}
	suite.SetUpSuite(c)
	suite.PatchValue(&dest, expSuiteVal)
	c.Assert(dest, gc.Equals, expSuiteVal)
	suite.SetUpTest(c)
	suite.PatchValue(&dest, expTestVal)
	c.Assert(dest, gc.Equals, expTestVal)
	suite.TearDownTest(c)
	suite.SetUpTest(c)
	c.Assert(dest, gc.Equals, expSuiteVal)
	suite.TearDownTest(c)
}

func (s *cleanupSuite) TestTearDownSuiteWithPatch(c *gc.C) {
	expSuiteVal := 42
	dest := 0
	suite := testing.CleanupSuite{}
	suite.SetUpSuite(c)
	suite.PatchValue(&dest, expSuiteVal)
	c.Assert(dest, gc.Equals, expSuiteVal)
	suite.TearDownSuite(c)
	c.Assert(dest, gc.Equals, 0)
}

func (s *cleanupSuite) TestAddSuiteCleanup(c *gc.C) {
	suite := testing.CleanupSuite{}
	order := []string{}
	suite.AddCleanup(func(*gc.C) {
		order = append(order, "first")
	})
	suite.AddCleanup(func(*gc.C) {
		order = append(order, "second")
	})

	suite.TearDownSuite(c)
	c.Assert(order, gc.DeepEquals, []string{"second", "first"})
}

func (s *cleanupSuite) TestAddCleanup(c *gc.C) {
	suite := testing.CleanupSuite{}
	order := []string{}
	goCheck := gc.C{}
	suite.SetUpTest(&goCheck)
	suite.AddCleanup(func(*gc.C) {
		order = append(order, "first")
	})
	suite.AddCleanup(func(*gc.C) {
		order = append(order, "second")
	})

	suite.TearDownTest(c)
	c.Assert(order, gc.DeepEquals, []string{"second", "first"})
}

func (s *cleanupSuite) TestPatchEnvironment(c *gc.C) {
	suite := testing.CleanupSuite{}
	goCheck := gc.C{}
	suite.SetUpTest(&goCheck)
	const envName = "TESTING_PATCH_ENVIRONMENT"
	os.Setenv(envName, "initial")

	suite.PatchEnvironment(envName, "new value")
	// Using check to make sure the environment gets set back properly in the test.
	c.Check(os.Getenv(envName), gc.Equals, "new value")

	suite.TearDownTest(&goCheck)
	c.Check(os.Getenv(envName), gc.Equals, "initial")
}

func (s *cleanupSuite) TestPatchValueInt(c *gc.C) {
	suite := testing.CleanupSuite{}
	goCheck := gc.C{}
	suite.SetUpTest(&goCheck)
	i := 42
	suite.PatchValue(&i, 0)
	c.Assert(i, gc.Equals, 0)

	suite.TearDownTest(c)
	c.Assert(i, gc.Equals, 42)
}

func (s *cleanupSuite) TestPatchValueFunction(c *gc.C) {
	suite := testing.CleanupSuite{}
	goCheck := gc.C{}
	suite.SetUpTest(&goCheck)
	function := func() string {
		return "original"
	}

	suite.PatchValue(&function, func() string {
		return "patched"
	})
	c.Assert(function(), gc.Equals, "patched")

	suite.TearDownTest(&goCheck)
	c.Assert(function(), gc.Equals, "original")
}
