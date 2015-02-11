// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

// FakeCallArgs holds all the args passed in a call.
type FakeCallArgs map[string]interface{}

// FakeCall records the name of a called function and the passed args.
type FakeCall struct {
	// Receiver is the fake for which the function was called. It is
	// not required, particularly if the function is not a method.
	Receiver interface{}

	// Funcname is the name of the function that was called.
	FuncName string

	// Args is the set of arguments to the function.
	Args FakeCallArgs
}

// Fake is used in testing to stand in for some other value, to record
// all calls to faked methods/functions, and to allow users to set the
// values that are returned from those calls. Fake is intended to be
// embedded in another struct that will define the methods to track:
//
//    type fakeConn struct {
//        *testing.Fake
//        Response []byte
//    }
//
//    func newFakeConn() *fakeConn {
//        return &fakeConn{
//            Fake: &testing.Fake{},
//        }
//    }
//
//    func (fc fakeConn) Send(request string) []byte {
//        fc.AddCall("Send", FakeCallArgs{
//            "request": request,
//        })
//        return fc.Response, fc.NextErr()
//    }
//
// As demonstrated in the example, embed a pointer to testing.Fake. This
// allows a single testing.Fake to be shared between multiple fakes.
//
// Error return values are set through Fake.Errors. Set it to the errors
// you want returned. The errors will be matched up to the calls made on
// Fake methods, in order. This is facilitated by the Err method, as
// seen in the above example. In some cases the first method call is not
// the one you want to fail. If not then put a nil before the error in
// Fake.Errors. Fake.SetErrors is a helper for setting up failure cases.
//
// To validate calls made to the fake in a test, call the CheckCalls
// method:
//
//    s.fake.CheckCalls(c, []FakeCall{{
//        FuncName: "Send",
//        Args: FakeCallArgs{
//            "request": expected,
//        },
//    }
//
// Not only is Fake useful for building a interface implementation to
// use in testing (e.g. a network API client), it is also useful in
// patching situations:
//
//    s.PatchValue(&somefunc, s.fake.Send)
//
// This allows for easily monitoring the args passed to the patched
// func, as well as controlling the return value from the func in a
// clean manner (by simply setting the correct fake field).
type Fake struct {
	// Calls is the list of calls that have been registered on the fake
	// (i.e. made on the fake's methods), in the order that they were
	// made.
	Calls []FakeCall

	// Errors holds the list of error return values to use for
	// successive calls to methods that return an error. Each call
	// pops the next error off the list. An empty list (the default)
	// implies a nil error. nil may be precede actual errors in the
	// list, which means that the first calls will succeed, followed
	// by the failure. All this is facilitated through the Err method.
	Errors []error

	// Error is the default error (when Errors is empty). The typical
	// Fake usage will leave this nil (i.e. no error).
	Error error
}

// NextErr returns the error that should be returned on the nth call to
// any method on the fake. It should be called for the error return in
// all faked methods.
func (f *Fake) NextErr() error {
	if len(f.Errors) == 0 {
		return f.Error
	}
	err := f.Errors[0]
	f.Errors = f.Errors[1:]
	return err
}

// AddCall records a faked method call for later inspection using the
// CheckCalls method. All faked methods should call AddCall.
func (f *Fake) AddCall(funcName string, args FakeCallArgs) {
	f.Calls = append(f.Calls, FakeCall{
		FuncName: funcName,
		Args:     args,
	})
}

// AddRcvrCall records a faked method call for later inspection using
// the CheckCalls method. All faked methods should call AddCall.
func (f *Fake) AddRcvrCall(receiver interface{}, funcName string, args FakeCallArgs) {
	f.Calls = append(f.Calls, FakeCall{
		Receiver: receiver,
		FuncName: funcName,
		Args:     args,
	})
}

// SetErrors sets the errors for the fake. Each call to Err (thus each
// fake method call) pops an error off the front. So frontloading nil
// here will allow calls to pass, followed by a failure.
func (f *Fake) SetErrors(errors ...error) {
	f.Errors = errors
}

// CheckCalls verifies that the history of calls on the fake's methods
// matches the expected calls.
func (f *Fake) CheckCalls(c *gc.C, expected []FakeCall) {
	c.Check(f.Calls, jc.DeepEquals, expected)
}

// CheckCallNames verifies that the in-order list of called method names
// matches the expected calls.
func (f *Fake) CheckCallNames(c *gc.C, expected ...string) {
	var funcNames []string
	for _, call := range f.Calls {
		funcNames = append(funcNames, call.FuncName)
	}
	c.Check(funcNames, jc.DeepEquals, expected)
}

// Reset sets the fake back to a pristine state.
func (f *Fake) Reset() {
	f.ResetCalls()
	f.Errors = nil
}

// ResetCalls clears the history of calls.
func (f *Fake) ResetCalls() {
	f.Calls = nil
}
