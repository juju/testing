// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing_test

import (
	"github.com/juju/errors"
	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
)

type fakeA struct {
	*testing.Fake
}

func (f *fakeA) aMethod(a, b, c int) error {
	f.MethodCall(f, "aMethod", a, b, c)
	return f.NextErr()
}

func (f *fakeA) otherMethod(values ...string) error {
	f.MethodCall(f, "otherMethod", values)
	return f.NextErr()
}

type fakeB struct {
	*testing.Fake
}

func (f *fakeB) aMethod() error {
	f.MethodCall(f, "aMethod")
	return f.NextErr()
}

func (f *fakeB) aFunc(value string) error {
	f.AddCall("aFunc", value)
	return f.NextErr()
}

type fakeSuite struct {
	fake *testing.Fake
}

var _ = gc.Suite(&fakeSuite{})

func (s *fakeSuite) SetUpTest(c *gc.C) {
	s.fake = &testing.Fake{}
}

func (s *fakeSuite) TestNextErrSequence(c *gc.C) {
	exp1 := errors.New("<failure 1>")
	exp2 := errors.New("<failure 2>")
	s.fake.Errors = []error{exp1, exp2}

	err1 := s.fake.NextErr()
	err2 := s.fake.NextErr()

	c.Check(err1, gc.Equals, exp1)
	c.Check(err2, gc.Equals, exp2)
}

func (s *fakeSuite) TestNextErrPops(c *gc.C) {
	exp1 := errors.New("<failure 1>")
	exp2 := errors.New("<failure 2>")
	s.fake.Errors = []error{exp1, exp2}

	s.fake.NextErr()

	c.Check(s.fake.Errors, jc.DeepEquals, []error{exp2})
}

func (s *fakeSuite) TestNextErrEmptyNil(c *gc.C) {
	err1 := s.fake.NextErr()
	err2 := s.fake.NextErr()

	c.Check(err1, jc.ErrorIsNil)
	c.Check(err2, jc.ErrorIsNil)
}

func (s *fakeSuite) TestNextErrDefault(c *gc.C) {
	expected := errors.New("<failure>")
	s.fake.DefaultError = expected

	err := s.fake.NextErr()

	c.Check(err, gc.Equals, expected)
}

func (s *fakeSuite) TestNextErrNil(c *gc.C) {
	s.fake.DefaultError = errors.New("<failure>")
	s.fake.Errors = []error{nil}

	err := s.fake.NextErr()

	c.Check(err, jc.ErrorIsNil)
}

func (s *fakeSuite) TestNextErrSkip(c *gc.C) {
	expected := errors.New("<failure>")
	s.fake.Errors = []error{nil, nil, expected}

	err1 := s.fake.NextErr()
	err2 := s.fake.NextErr()
	err3 := s.fake.NextErr()

	c.Check(err1, jc.ErrorIsNil)
	c.Check(err2, jc.ErrorIsNil)
	c.Check(err3, gc.Equals, expected)
}

func (s *fakeSuite) TestNextErrEmbeddedMixed(c *gc.C) {
	exp1 := errors.New("<failure 1>")
	exp2 := errors.New("<failure 2>")
	s.fake.Errors = []error{exp1, nil, nil, exp2}

	fake1 := &fakeA{s.fake}
	fake2 := &fakeB{s.fake}
	err1 := fake1.aMethod(1, 2, 3)
	err2 := fake2.aFunc("arg")
	err3 := fake1.otherMethod("arg1", "arg2")
	err4 := fake2.aMethod()

	c.Check(err1, gc.Equals, exp1)
	c.Check(err2, jc.ErrorIsNil)
	c.Check(err3, jc.ErrorIsNil)
	c.Check(err4, gc.Equals, exp2)
}

func (s *fakeSuite) TestAddCallRecorded(c *gc.C) {
	s.fake.AddCall("aFunc", 1, 2, 3)

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		FuncName: "aFunc",
		Args:     []interface{}{1, 2, 3},
	}})
}

func (s *fakeSuite) TestAddCallRepeated(c *gc.C) {
	s.fake.AddCall("before", "arg")
	s.fake.AddCall("aFunc", 1, 2, 3)
	s.fake.AddCall("aFunc", 4, 5, 6)
	s.fake.AddCall("after", "arg")

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		FuncName: "before",
		Args:     []interface{}{"arg"},
	}, {
		FuncName: "aFunc",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "aFunc",
		Args:     []interface{}{4, 5, 6},
	}, {
		FuncName: "after",
		Args:     []interface{}{"arg"},
	}})
}

func (s *fakeSuite) TestAddCallNoArgs(c *gc.C) {
	s.fake.AddCall("aFunc")

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		FuncName: "aFunc",
	}})
}

func (s *fakeSuite) TestAddCallSequence(c *gc.C) {
	s.fake.AddCall("first")
	s.fake.AddCall("second")
	s.fake.AddCall("third")

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		FuncName: "first",
	}, {
		FuncName: "second",
	}, {
		FuncName: "third",
	}})
}

func (s *fakeSuite) TestMethodCall(c *gc.C) {
	s.fake.MethodCall(s.fake, "aMethod", 1, 2, 3)

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		Receiver: s.fake,
		FuncName: "aMethod",
		Args:     []interface{}{1, 2, 3},
	}})
}

func (s *fakeSuite) TestMethodCallMixed(c *gc.C) {
	s.fake.MethodCall(s.fake, "Method1", 1, 2, 3)
	s.fake.AddCall("aFunc", "arg")
	s.fake.MethodCall(s.fake, "Method2")

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		Receiver: s.fake,
		FuncName: "Method1",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "aFunc",
		Args:     []interface{}{"arg"},
	}, {
		Receiver: s.fake,
		FuncName: "Method2",
	}})
}

func (s *fakeSuite) TestMethodCallEmbeddedMixed(c *gc.C) {
	fake1 := &fakeA{s.fake}
	fake2 := &fakeB{s.fake}
	err := fake1.aMethod(1, 2, 3)
	c.Assert(err, jc.ErrorIsNil)
	err = fake2.aFunc("arg")
	c.Assert(err, jc.ErrorIsNil)
	err = fake1.otherMethod("arg1", "arg2")
	c.Assert(err, jc.ErrorIsNil)
	err = fake2.aMethod()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.fake.Calls, jc.DeepEquals, []testing.FakeCall{{
		Receiver: fake1,
		FuncName: "aMethod",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "aFunc",
		Args:     []interface{}{"arg"},
	}, {
		Receiver: fake1,
		FuncName: "otherMethod",
		Args:     []interface{}{[]string{"arg1", "arg2"}},
	}, {
		Receiver: fake2,
		FuncName: "aMethod",
	}})
}

func (s *fakeSuite) TestSetErrorsMultiple(c *gc.C) {
	err1 := errors.New("<failure 1>")
	err2 := errors.New("<failure 2>")
	s.fake.SetErrors(err1, err2)

	c.Check(s.fake.Errors, jc.DeepEquals, []error{err1, err2})
}

func (s *fakeSuite) TestSetErrorsEmpty(c *gc.C) {
	s.fake.SetErrors()

	c.Check(s.fake.Errors, gc.HasLen, 0)
}

func (s *fakeSuite) TestSetErrorMixed(c *gc.C) {
	err1 := errors.New("<failure 1>")
	err2 := errors.New("<failure 2>")
	s.fake.SetErrors(nil, err1, nil, err2)

	c.Check(s.fake.Errors, jc.DeepEquals, []error{nil, err1, nil, err2})
}

func (s *fakeSuite) TestSetErrorsTrailingNil(c *gc.C) {
	err := errors.New("<failure 1>")
	s.fake.SetErrors(err, nil)

	c.Check(s.fake.Errors, jc.DeepEquals, []error{err, nil})
}

func (s *fakeSuite) checkCallsStandard(c *gc.C) {
	s.fake.CheckCalls(c, []testing.FakeCall{{
		FuncName: "first",
		Args:     []interface{}{"arg"},
	}, {
		FuncName: "second",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "third",
	}})
}

func (s *fakeSuite) TestCheckCallsPass(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("second", 1, 2, 3)
	s.fake.AddCall("third")

	s.checkCallsStandard(c)
}

func (s *fakeSuite) TestCheckCallsEmpty(c *gc.C) {
	s.fake.CheckCalls(c, nil)
}

func (s *fakeSuite) TestCheckCallsMissingCall(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCalls call should fail`)
	s.checkCallsStandard(c)
}

func (s *fakeSuite) TestCheckCallsWrongName(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("oops", 1, 2, 3)
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCalls call should fail`)
	s.checkCallsStandard(c)
}

func (s *fakeSuite) TestCheckCallsWrongArgs(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("second", 1, 2, 4)
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCalls call should fail`)
	s.checkCallsStandard(c)
}

func (s *fakeSuite) checkCallStandard(c *gc.C) {
	s.fake.CheckCall(c, 0, "first", "arg")
	s.fake.CheckCall(c, 1, "second", 1, 2, 3)
	s.fake.CheckCall(c, 2, "third")
}

func (s *fakeSuite) TestCheckCallPass(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("second", 1, 2, 3)
	s.fake.AddCall("third")

	s.checkCallStandard(c)
}

func (s *fakeSuite) TestCheckCallEmpty(c *gc.C) {
	c.ExpectFailure(`Fake.CheckCall should fail when no calls have been made`)
	s.fake.CheckCall(c, 0, "aMethod")
}

func (s *fakeSuite) TestCheckCallMissingCall(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCall call should fail here`)
	s.checkCallStandard(c)
}

func (s *fakeSuite) TestCheckCallWrongName(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("oops", 1, 2, 3)
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCall call should fail here`)
	s.checkCallStandard(c)
}

func (s *fakeSuite) TestCheckCallWrongArgs(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("second", 1, 2, 4)
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCall call should fail here`)
	s.checkCallStandard(c)
}

func (s *fakeSuite) TestCheckCallNamesPass(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("second", 1, 2, 4)
	s.fake.AddCall("third")

	s.fake.CheckCallNames(c, "first", "second", "third")
}

func (s *fakeSuite) TestCheckCallNamesUnexpected(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("second", 1, 2, 4)
	s.fake.AddCall("third")

	c.ExpectFailure(`Fake.CheckCall should fail when no calls have been made`)
	s.fake.CheckCallNames(c)
}

func (s *fakeSuite) TestCheckCallNamesEmptyPass(c *gc.C) {
	s.fake.CheckCallNames(c)
}

func (s *fakeSuite) TestCheckCallNamesEmptyFail(c *gc.C) {
	c.ExpectFailure(`Fake.CheckCall should fail when no calls have been made`)
	s.fake.CheckCallNames(c, "aMethod")
}

func (s *fakeSuite) TestCheckCallNamesMissingCall(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCallNames call should fail here`)
	s.fake.CheckCallNames(c, "first", "second", "third")
}

func (s *fakeSuite) TestCheckCallNamesWrongName(c *gc.C) {
	s.fake.AddCall("first", "arg")
	s.fake.AddCall("oops", 1, 2, 4)
	s.fake.AddCall("third")

	c.ExpectFailure(`the "standard" Fake.CheckCallNames call should fail here`)
	s.fake.CheckCallNames(c, "first", "second", "third")
}
