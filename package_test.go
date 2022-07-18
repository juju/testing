// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	stdtesting "testing"

	gc "gopkg.in/check.v1"
)

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}
