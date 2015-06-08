// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

// Suite registers the provided suite with the test runner, but only
// if the "default" test tags were passed at the commandline, which they
// are by default. Use Suite in place of gc.Suite.
func Suite(suite interface{}) {
	SuiteTagged(suite, defaultTags...)
}
