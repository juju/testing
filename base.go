// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

// Suite registers the provided suite with the test runner, but only
// if the base test tags was passed at the commandline, which it is
// by default. Use Suite in place of gc.Suite.
func Suite(suite interface{}) {
	SuiteTagged(suite, TagBase)
}
