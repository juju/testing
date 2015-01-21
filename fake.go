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
	FuncName string
	Args     FakeCallArgs
}

// Fake is used in testing to stand in for some other value, to record
// all calls to faked methods/functions, and to allow users to set the
// values that are returned from those calls. Fake is intended to be
// embedded in another struct that will define the methods to track:
//
//    type fakeConn struct {
//        fake
//        Response []byte
//    }
//
//    func (fc fakeConn) Send(request string) []byte {
//        fc.AddCall("Send", FakeCallArgs{
//            "request": request,
//        })
//        return fc.Response, fc.Error()
//    }
//
// Fake has two fields for error situations. Setting "Err" is intended
// to cause any method call to fail. This is facilitated by the Error
// method, as seen in the above example. In some cases the first method
// call is not the one you want to fail. If not then set "FailOnCall" to
// the ordered index of the call which should fail.
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
	calls []FakeCall

	Err        error
	FailOnCall int
}

// Error returns the error that should be returned on the nth call to
// any method on the fake. It should be called for the error return in
// all faked methods.
func (f *Fake) Error() error {
	if len(f.calls) != f.FailOnCall+1 {
		return nil
	}
	return f.Err
}

// AddCall records a faked method call for later inspection using the
// CheckCalls method. All faked methods should call AddCall.
func (f *Fake) AddCall(funcName string, args FakeCallArgs) {
	f.calls = append(f.calls, FakeCall{
		FuncName: funcName,
		Args:     args,
	})
}

// CheckCalls verifies that the history of calls on the fake's methods
// matches the expected calls.
func (f *Fake) CheckCalls(c *gc.C, expected []FakeCall) {
	c.Check(f.calls, jc.DeepEquals, expected)
}
