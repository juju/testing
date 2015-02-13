// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing_test

import (
	"github.com/juju/errors"
	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
)

type mockA struct {
	*testing.Mock
}

func (f *mockA) aMethod(a, b, c int) error {
	f.MethodCall(f, "aMethod", a, b, c)
	return f.NextErr()
}

func (f *mockA) otherMethod(values ...string) error {
	f.MethodCall(f, "otherMethod", values)
	return f.NextErr()
}

type mockB struct {
	*testing.Mock
}

func (f *mockB) aMethod() error {
	f.MethodCall(f, "aMethod")
	return f.NextErr()
}

func (f *mockB) aFunc(value string) error {
	f.AddCall("aFunc", value)
	return f.NextErr()
}

type mockSuite struct {
	mock *testing.Mock
}

var _ = gc.Suite(&mockSuite{})

func (s *mockSuite) SetUpTest(c *gc.C) {
	s.mock = &testing.Mock{}
}

func (s *mockSuite) TestNextErrSequence(c *gc.C) {
	exp1 := errors.New("<failure 1>")
	exp2 := errors.New("<failure 2>")
	s.mock.Errors = []error{exp1, exp2}

	err1 := s.mock.NextErr()
	err2 := s.mock.NextErr()

	c.Check(err1, gc.Equals, exp1)
	c.Check(err2, gc.Equals, exp2)
}

func (s *mockSuite) TestNextErrPops(c *gc.C) {
	exp1 := errors.New("<failure 1>")
	exp2 := errors.New("<failure 2>")
	s.mock.Errors = []error{exp1, exp2}

	s.mock.NextErr()

	c.Check(s.mock.Errors, jc.DeepEquals, []error{exp2})
}

func (s *mockSuite) TestNextErrEmptyNil(c *gc.C) {
	err1 := s.mock.NextErr()
	err2 := s.mock.NextErr()

	c.Check(err1, jc.ErrorIsNil)
	c.Check(err2, jc.ErrorIsNil)
}

func (s *mockSuite) TestNextErrDefault(c *gc.C) {
	expected := errors.New("<failure>")
	s.mock.DefaultError = expected

	err := s.mock.NextErr()

	c.Check(err, gc.Equals, expected)
}

func (s *mockSuite) TestNextErrNil(c *gc.C) {
	s.mock.DefaultError = errors.New("<failure>")
	s.mock.Errors = []error{nil}

	err := s.mock.NextErr()

	c.Check(err, jc.ErrorIsNil)
}

func (s *mockSuite) TestNextErrSkip(c *gc.C) {
	expected := errors.New("<failure>")
	s.mock.Errors = []error{nil, nil, expected}

	err1 := s.mock.NextErr()
	err2 := s.mock.NextErr()
	err3 := s.mock.NextErr()

	c.Check(err1, jc.ErrorIsNil)
	c.Check(err2, jc.ErrorIsNil)
	c.Check(err3, gc.Equals, expected)
}

func (s *mockSuite) TestNextErrEmbeddedMixed(c *gc.C) {
	exp1 := errors.New("<failure 1>")
	exp2 := errors.New("<failure 2>")
	s.mock.Errors = []error{exp1, nil, nil, exp2}

	mock1 := &mockA{s.mock}
	mock2 := &mockB{s.mock}
	err1 := mock1.aMethod(1, 2, 3)
	err2 := mock2.aFunc("arg")
	err3 := mock1.otherMethod("arg1", "arg2")
	err4 := mock2.aMethod()

	c.Check(err1, gc.Equals, exp1)
	c.Check(err2, jc.ErrorIsNil)
	c.Check(err3, jc.ErrorIsNil)
	c.Check(err4, gc.Equals, exp2)
}

func (s *mockSuite) TestAddCallRecorded(c *gc.C) {
	s.mock.AddCall("aFunc", 1, 2, 3)

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
		FuncName: "aFunc",
		Args:     []interface{}{1, 2, 3},
	}})
	c.Check(s.mock.Receivers, jc.DeepEquals, []interface{}{nil})
}

func (s *mockSuite) TestAddCallRepeated(c *gc.C) {
	s.mock.AddCall("before", "arg")
	s.mock.AddCall("aFunc", 1, 2, 3)
	s.mock.AddCall("aFunc", 4, 5, 6)
	s.mock.AddCall("after", "arg")

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
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
	c.Check(s.mock.Receivers, jc.DeepEquals, []interface{}{nil, nil, nil, nil})
}

func (s *mockSuite) TestAddCallNoArgs(c *gc.C) {
	s.mock.AddCall("aFunc")

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
		FuncName: "aFunc",
	}})
}

func (s *mockSuite) TestAddCallSequence(c *gc.C) {
	s.mock.AddCall("first")
	s.mock.AddCall("second")
	s.mock.AddCall("third")

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
		FuncName: "first",
	}, {
		FuncName: "second",
	}, {
		FuncName: "third",
	}})
}

func (s *mockSuite) TestMethodCallRecorded(c *gc.C) {
	s.mock.MethodCall(s.mock, "aMethod", 1, 2, 3)

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
		FuncName: "aMethod",
		Args:     []interface{}{1, 2, 3},
	}})
	c.Check(s.mock.Receivers, jc.DeepEquals, []interface{}{s.mock})
}

func (s *mockSuite) TestMethodCallMixed(c *gc.C) {
	s.mock.MethodCall(s.mock, "Method1", 1, 2, 3)
	s.mock.AddCall("aFunc", "arg")
	s.mock.MethodCall(s.mock, "Method2")

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
		FuncName: "Method1",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "aFunc",
		Args:     []interface{}{"arg"},
	}, {
		FuncName: "Method2",
	}})
	c.Check(s.mock.Receivers, jc.DeepEquals, []interface{}{s.mock, nil, s.mock})
}

func (s *mockSuite) TestMethodCallEmbeddedMixed(c *gc.C) {
	mock1 := &mockA{s.mock}
	mock2 := &mockB{s.mock}
	err := mock1.aMethod(1, 2, 3)
	c.Assert(err, jc.ErrorIsNil)
	err = mock2.aFunc("arg")
	c.Assert(err, jc.ErrorIsNil)
	err = mock1.otherMethod("arg1", "arg2")
	c.Assert(err, jc.ErrorIsNil)
	err = mock2.aMethod()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.mock.Calls, jc.DeepEquals, []testing.MockCall{{
		FuncName: "aMethod",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "aFunc",
		Args:     []interface{}{"arg"},
	}, {
		FuncName: "otherMethod",
		Args:     []interface{}{[]string{"arg1", "arg2"}},
	}, {
		FuncName: "aMethod",
	}})
	c.Check(s.mock.Receivers, jc.DeepEquals, []interface{}{mock1, nil, mock1, mock2})
}

func (s *mockSuite) TestSetErrorsMultiple(c *gc.C) {
	err1 := errors.New("<failure 1>")
	err2 := errors.New("<failure 2>")
	s.mock.SetErrors(err1, err2)

	c.Check(s.mock.Errors, jc.DeepEquals, []error{err1, err2})
}

func (s *mockSuite) TestSetErrorsEmpty(c *gc.C) {
	s.mock.SetErrors()

	c.Check(s.mock.Errors, gc.HasLen, 0)
}

func (s *mockSuite) TestSetErrorMixed(c *gc.C) {
	err1 := errors.New("<failure 1>")
	err2 := errors.New("<failure 2>")
	s.mock.SetErrors(nil, err1, nil, err2)

	c.Check(s.mock.Errors, jc.DeepEquals, []error{nil, err1, nil, err2})
}

func (s *mockSuite) TestSetErrorsTrailingNil(c *gc.C) {
	err := errors.New("<failure 1>")
	s.mock.SetErrors(err, nil)

	c.Check(s.mock.Errors, jc.DeepEquals, []error{err, nil})
}

func (s *mockSuite) checkCallsStandard(c *gc.C) {
	s.mock.CheckCalls(c, []testing.MockCall{{
		FuncName: "first",
		Args:     []interface{}{"arg"},
	}, {
		FuncName: "second",
		Args:     []interface{}{1, 2, 3},
	}, {
		FuncName: "third",
	}})
}

func (s *mockSuite) TestCheckCallsPass(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("second", 1, 2, 3)
	s.mock.AddCall("third")

	s.checkCallsStandard(c)
}

func (s *mockSuite) TestCheckCallsEmpty(c *gc.C) {
	s.mock.CheckCalls(c, nil)
}

func (s *mockSuite) TestCheckCallsMissingCall(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCalls call should fail`)
	s.checkCallsStandard(c)
}

func (s *mockSuite) TestCheckCallsWrongName(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("oops", 1, 2, 3)
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCalls call should fail`)
	s.checkCallsStandard(c)
}

func (s *mockSuite) TestCheckCallsWrongArgs(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("second", 1, 2, 4)
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCalls call should fail`)
	s.checkCallsStandard(c)
}

func (s *mockSuite) checkCallStandard(c *gc.C) {
	s.mock.CheckCall(c, 0, "first", "arg")
	s.mock.CheckCall(c, 1, "second", 1, 2, 3)
	s.mock.CheckCall(c, 2, "third")
}

func (s *mockSuite) TestCheckCallPass(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("second", 1, 2, 3)
	s.mock.AddCall("third")

	s.checkCallStandard(c)
}

func (s *mockSuite) TestCheckCallEmpty(c *gc.C) {
	c.ExpectFailure(`Mock.CheckCall should fail when no calls have been made`)
	s.mock.CheckCall(c, 0, "aMethod")
}

func (s *mockSuite) TestCheckCallMissingCall(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCall call should fail here`)
	s.checkCallStandard(c)
}

func (s *mockSuite) TestCheckCallWrongName(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("oops", 1, 2, 3)
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCall call should fail here`)
	s.checkCallStandard(c)
}

func (s *mockSuite) TestCheckCallWrongArgs(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("second", 1, 2, 4)
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCall call should fail here`)
	s.checkCallStandard(c)
}

func (s *mockSuite) TestCheckCallNamesPass(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("second", 1, 2, 4)
	s.mock.AddCall("third")

	s.mock.CheckCallNames(c, "first", "second", "third")
}

func (s *mockSuite) TestCheckCallNamesUnexpected(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("second", 1, 2, 4)
	s.mock.AddCall("third")

	c.ExpectFailure(`Mock.CheckCall should fail when no calls have been made`)
	s.mock.CheckCallNames(c)
}

func (s *mockSuite) TestCheckCallNamesEmptyPass(c *gc.C) {
	s.mock.CheckCallNames(c)
}

func (s *mockSuite) TestCheckCallNamesEmptyFail(c *gc.C) {
	c.ExpectFailure(`Mock.CheckCall should fail when no calls have been made`)
	s.mock.CheckCallNames(c, "aMethod")
}

func (s *mockSuite) TestCheckCallNamesMissingCall(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCallNames call should fail here`)
	s.mock.CheckCallNames(c, "first", "second", "third")
}

func (s *mockSuite) TestCheckCallNamesWrongName(c *gc.C) {
	s.mock.AddCall("first", "arg")
	s.mock.AddCall("oops", 1, 2, 4)
	s.mock.AddCall("third")

	c.ExpectFailure(`the "standard" Mock.CheckCallNames call should fail here`)
	s.mock.CheckCallNames(c, "first", "second", "third")
}
