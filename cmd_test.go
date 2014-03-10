// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"os/exec"

	gc "launchpad.net/gocheck"

	"github.com/juju/testing"
)

type cmdSuite struct{}

var _ = gc.Suite(&cmdSuite{})

func (s *cmdSuite) TestHookCommandOutput(c *gc.C) {
	var CommandOutput = (*exec.Cmd).CombinedOutput

	cmdChan, cleanup := testing.HookCommandOutput(&CommandOutput, []byte{1, 2, 3, 4}, nil)
	defer cleanup()

	testCmd := exec.Command("fake-command", "arg1", "arg2")
	out, err := CommandOutput(testCmd)
	c.Assert(err, gc.IsNil)
	cmd := <-cmdChan
	c.Assert(out, gc.DeepEquals, []byte{1, 2, 3, 4})
	c.Assert(cmd.Args, gc.DeepEquals, []string{"fake-command", "arg1", "arg2"})
}
