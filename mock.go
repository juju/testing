// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

// MockCall records the name of a called function and the passed args.
type MockCall struct {
	// Funcname is the name of the function that was called.
	FuncName string

	// Args is the set of arguments passed to the function. They are
	// in the same order as the function's parameters
	Args []interface{}
}

// Mock is used in testing to stand in for some other value, to record
// all calls to mocked methods/functions, and to allow users to set the
// values that are returned from those calls. Mock is intended to be
// embedded in another struct that will define the methods to track:
//
//    type mockConn struct {
//        *testing.Mock
//        Response []byte
//    }
//
//    func newMockConn() *mockConn {
//        return &mockConn{
//            Mock: &testing.Mock{},
//        }
//    }
//
//    // Send implements Connection.
//    func (fc *mockConn) Send(request string) []byte {
//        fc.MethodCall(fc, "Send", request)
//        return fc.Response, fc.NextErr()
//    }
//
// As demonstrated in the example, embed a pointer to testing.Mock. This
// allows a single testing.Mock to be shared between multiple mocks.
//
// Error return values are set through Mock.Errors. Set it to the errors
// you want returned (or use the convenience method `SetErrors`). The
// `NextErr` method returns the errors from Mock.Errors in sequence,
// falling back to `DefaultError` when the sequence is exhausted. Thus
// each mocked method should call `NextErr` to get its error return value.
//
// To validate calls made to the mock in a test, check Mock.Calls or
// call the CheckCalls (or CheckCall) method:
//
//    c.Check(s.mock.Calls, jc.DeepEquals, []MockCall{{
//        FuncName: "Send",
//        Args: []interface{}{
//            expected,
//        },
//    }})
//
//    s.mock.CheckCalls(c, []MockCall{{
//        FuncName: "Send",
//        Args: []interface{}{
//            expected,
//        },
//    }})
//
//    s.mock.CheckCall(c, 0, "Send", expected)
//
// Not only is Mock useful for building a interface implementation to
// use in testing (e.g. a network API client), it is also useful in
// regular function patching situations:
//
//    type myMock struct {
//        *testing.Mock
//    }
//
//    func (f *myMock) SomeFunc(arg interface{}) error {
//        f.AddCall("SomeFunc", arg)
//        return f.NextErr()
//    }
//
//    s.PatchValue(&somefunc, s.myMock.SomeFunc)
//
// This allows for easily monitoring the args passed to the patched
// func, as well as controlling the return value from the func in a
// clean manner (by simply setting the correct field on the mock).
type Mock struct {
	// Calls is the list of calls that have been registered on the mock
	// (i.e. made on the mock's methods), in the order that they were
	// made.
	Calls []MockCall

	// Receivers is the list of receivers for all the recorded calls.
	// In the case of non-methods, the receiver is set to nil. The
	// receivers are tracked here rather than as a Receiver field on
	// MockCall because MockCall represents the common case for
	// testing. Typically the receiver does not need to be checked.
	Receivers []interface{}

	// Errors holds the list of error return values to use for
	// successive calls to methods that return an error. Each call
	// pops the next error off the list. An empty list (the default)
	// implies a nil error. nil may be precede actual errors in the
	// list, which means that the first calls will succeed, followed
	// by the failure. All this is facilitated through the Err method.
	Errors []error

	// DefaultError is the default error (when Errors is empty). The
	// typical Mock usage will leave this nil (i.e. no error).
	DefaultError error
}

// TODO(ericsnow) Add something similar to NextErr for all return values
// using reflection?

// NextErr returns the error that should be returned on the nth call to
// any method on the mock. It should be called for the error return in
// all mocked methods.
func (f *Mock) NextErr() error {
	if len(f.Errors) == 0 {
		return f.DefaultError
	}
	err := f.Errors[0]
	f.Errors = f.Errors[1:]
	return err
}

func (f *Mock) addCall(rcvr interface{}, funcName string, args []interface{}) {
	f.Calls = append(f.Calls, MockCall{
		FuncName: funcName,
		Args:     args,
	})
	f.Receivers = append(f.Receivers, rcvr)
}

// AddCall records a mocked function call for later inspection using the
// CheckCalls method. A nil receiver is recorded. Thus for methods use
// MethodCall. All mocked functions should call AddCall.
func (f *Mock) AddCall(funcName string, args ...interface{}) {
	f.addCall(nil, funcName, args)
}

// MethodCall records a mocked method call for later inspection using
// the CheckCalls method. The receiver is added to Mock.Receivers.
func (f *Mock) MethodCall(receiver interface{}, funcName string, args ...interface{}) {
	f.addCall(receiver, funcName, args)
}

// SetErrors sets the sequence of error returns for the mock. Each call
// to Err (thus each mock method call) pops an error off the front. So
// frontloading nil here will allow calls to pass, followed by a
// failure.
func (f *Mock) SetErrors(errors ...error) {
	f.Errors = errors
}

// CheckCalls verifies that the history of calls on the mock's methods
// matches the expected calls. The receivers are not checked. If they
// are significant then check Mock.Receivers separately.
func (f *Mock) CheckCalls(c *gc.C, expected []MockCall) {
	if !f.CheckCallNames(c, mockCallNames(expected...)...) {
		return
	}
	c.Check(f.Calls, jc.DeepEquals, expected)
}

// CheckCall checks the recorded call at the given index against the
// provided values. If the index is out of bounds then the check fails.
// The receiver is not checked. If it is significant for a test then it
// can be checked separately:
//
//     c.Check(mymock.Receivers[index], gc.Equals, expected)
func (f *Mock) CheckCall(c *gc.C, index int, funcName string, args ...interface{}) {
	if !c.Check(index, jc.LessThan, len(f.Calls)) {
		return
	}
	call := f.Calls[index]
	expected := MockCall{
		FuncName: funcName,
		Args:     args,
	}
	c.Check(call, jc.DeepEquals, expected)
}

// CheckCallNames verifies that the in-order list of called method names
// matches the expected calls.
func (f *Mock) CheckCallNames(c *gc.C, expected ...string) bool {
	funcNames := mockCallNames(f.Calls...)
	return c.Check(funcNames, jc.DeepEquals, expected)
}

func mockCallNames(calls ...MockCall) []string {
	var funcNames []string
	for _, call := range calls {
		funcNames = append(funcNames, call.FuncName)
	}
	return funcNames
}
