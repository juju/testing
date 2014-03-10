// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"os/exec"

	gc "launchpad.net/gocheck"

	"github.com/juju/testing"
)

type cmdSuite struct {
	testing.CleanupSuite
}

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

func (s *cmdSuite) TestPatchExecutableNoArgs(c *gc.C) {
	c.Log(testing.EchoQuotedArgs)
	testing.PatchExecutable(c, s, "test-output", testing.EchoQuotedArgs)
	output := runCommand(c, "test-output")
	c.Assert(output, gc.Equals, "test-output\n")
	testing.AssertEchoArgs(c, "test-output")
}

func (s *cmdSuite) TestPatchExecutableWithArgs(c *gc.C) {
	testing.PatchExecutable(c, s, "test-output", testing.EchoQuotedArgs)
	output := runCommand(c, "test-output", "foo", "bar baz")
	c.Assert(output, gc.Equals, "test-output \"foo\" \"bar baz\"\n")
	testing.AssertEchoArgs(c, "test-output", "foo", "bar baz")
}

func runCommand(c *gc.C, command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	c.Assert(err, gc.IsNil)
	return string(out)
}
