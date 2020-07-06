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

type pod struct {
	A int
	a int
	B bool
	b bool
	C string
	c string
}

func (s *MultiCheckerSuite) TestPOD(c *gc.C) {
	a1 := pod{1, 2, true, true, "a", "a"}
	a2 := pod{2, 3, false, false, "b", "b"}

	checker := jc.NewMultiChecker().
		Add(".A", jc.Ignore).
		Add(".a", jc.Ignore).
		Add(".B", jc.Ignore).
		Add(".b", jc.Ignore).
		Add(".C", jc.Ignore).
		Add(".c", jc.Ignore)
	c.Check(a1, checker, a2)
}

func (s *MultiCheckerSuite) TestExprMap(c *gc.C) {
	a1 := map[string]string{"a": "a", "b": "b", "c": "c"}
	a2 := map[string]string{"a": "aaaa", "b": "bbbb", "c": "cccc"}

	checker := jc.NewMultiChecker().AddExpr(`_[_]`, jc.Ignore)
	c.Check(a1, checker, a2)
}

type complexA struct {
	complexB
	A int
	C []int
	D map[string]string
	E *complexE
	F **complexF
}

type complexB struct {
	B string
	b string
}

type complexE struct {
	E string
}

type complexF struct {
	F []string
}

func (s *MultiCheckerSuite) TestExprComplex(c *gc.C) {
	f1 := &complexF{
		F: []string{"a", "b"},
	}
	a1 := complexA{
		complexB: complexB{
			B: "wow",
			b: "wow",
		},
		A: 5,
		C: []int{0, 1, 2, 3, 4, 5},
		D: map[string]string{"a": "b"},
		E: &complexE{E: "E"},
		F: &f1,
	}
	f2 := &complexF{
		F: []string{"c", "d"},
	}
	a2 := complexA{
		complexB: complexB{
			B: "cool",
			b: "cool",
		},
		A: 19,
		C: []int{5, 4, 3, 2, 1, 0},
		D: map[string]string{"b": "a"},
		E: &complexE{E: "EEEEEEEEE"},
		F: &f2,
	}
	checker := jc.NewMultiChecker().
		AddExpr(`_.complexB.B`, jc.Ignore).
		AddExpr(`_.complexB.b`, jc.Ignore).
		AddExpr(`_.A`, jc.Ignore).
		AddExpr(`_.C[_]`, jc.Ignore).
		AddExpr(`_.D`, jc.Ignore).
		AddExpr(`(*_.E)`, jc.Ignore).
		AddExpr(`(*(*_.F)).F[_]`, jc.Ignore)
	c.Check(a1, checker, a2)
}
