package checkers_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type MultiCheckerSuite struct{}

var _ = gc.Suite(&MultiCheckerSuite{})

func (s *MultiCheckerSuite) TestDeepEquals(c *gc.C) {
	for i, test := range deepEqualTests {
		c.Logf("test %d. %v == %v is %v", i, test.a, test.b, test.eq)
		result, msg := jc.NewMultiChecker().Check([]interface{}{test.a, test.b}, nil)
		c.Check(result, gc.Equals, test.eq)
		if test.eq {
			c.Check(msg, gc.Equals, "")
		} else {
			c.Check(msg, gc.Not(gc.Equals), "")
		}
	}
}

func (s *MultiCheckerSuite) TestArray(c *gc.C) {
	a1 := []string{"a", "b", "c"}
	a2 := []string{"a", "bbb", "c"}

	checker := jc.NewMultiChecker().Add("[1]", jc.Ignore)
	c.Check(a1, checker, a2)
}

func (s *MultiCheckerSuite) TestMap(c *gc.C) {
	a1 := map[string]string{"a": "a", "b": "b", "c": "c"}
	a2 := map[string]string{"a": "a", "b": "bbbb", "c": "c"}

	checker := jc.NewMultiChecker().Add(`["b"]`, jc.Ignore)
	c.Check(a1, checker, a2)
}

func (s *MultiCheckerSuite) TestRegexArray(c *gc.C) {
	a1 := []string{"a", "b", "c"}
	a2 := []string{"a", "bbb", "ccc"}

	checker := jc.NewMultiChecker().AddRegex("\\[[1-2]\\]", jc.Ignore)
	c.Check(a1, checker, a2)
}

func (s *MultiCheckerSuite) TestRegexMap(c *gc.C) {
	a1 := map[string]string{"a": "a", "b": "b", "c": "c"}
	a2 := map[string]string{"a": "aaaa", "b": "bbbb", "c": "cccc"}

	checker := jc.NewMultiChecker().AddRegex(`\[".*"\]`, jc.Ignore)
	c.Check(a1, checker, a2)
}

func (s *MultiCheckerSuite) TestArrayArraysUnordered(c *gc.C) {
	a1 := [][]string{{"a", "b", "c"}, {"c", "d", "e"}}
	a2 := [][]string{{"a", "b", "c"}, {}}

	checker := jc.NewMultiChecker().Add("[1]", jc.SameContents, []string{"e", "c", "d"})
	c.Check(a1, checker, a2)
}

func (s *MultiCheckerSuite) TestArrayArraysUnorderedWithExpected(c *gc.C) {
	a1 := [][]string{{"a", "b", "c"}, {"c", "d", "e"}}
	a2 := [][]string{{"a", "b", "c"}, {"e", "c", "d"}}

	checker := jc.NewMultiChecker().Add("[1]", jc.SameContents, jc.ExpectedValue)
	c.Check(a1, checker, a2)
}
