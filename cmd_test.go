// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"os/exec"
	"strings"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
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

func (s *cmdSuite) EnsureArgFileRemoved(name string) {
	s.AddCleanup(func(c *gc.C) {
		c.Assert(name+".out", jc.DoesNotExist)
	})
}

const testFunc = "test-output"

func (s *cmdSuite) TestPatchExecutableNoArgs(c *gc.C) {
	s.EnsureArgFileRemoved(testFunc)
	testing.PatchExecutableAsEchoArgs(c, s, testFunc)
	output := runCommand(c, testFunc)
	output = strings.TrimRight(output, "\r\n")
	c.Assert(output, gc.Equals, testFunc)
	testing.AssertEchoArgs(c, testFunc)
}

func (s *cmdSuite) TestPatchExecutableWithArgs(c *gc.C) {
	s.EnsureArgFileRemoved(testFunc)
	testing.PatchExecutableAsEchoArgs(c, s, testFunc)
	output := runCommand(c, testFunc, "foo", "bar baz")
	output = strings.TrimRight(output, "\r\n")

	c.Assert(output, gc.DeepEquals, testFunc+" 'foo' 'bar baz'")

	testing.AssertEchoArgs(c, testFunc, "foo", "bar baz")
}

func (s *cmdSuite) TestPatchExecutableThrowError(c *gc.C) {
	testing.PatchExecutableThrowError(c, s, testFunc, 1)
	cmd := exec.Command(testFunc)
	out, err := cmd.CombinedOutput()
	c.Assert(err, gc.ErrorMatches, "exit status 1")
	output := strings.TrimRight(string(out), "\r\n")
	c.Assert(output, gc.Equals, "failing")
}

func runCommand(c *gc.C, command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	c.Assert(err, gc.IsNil)
	return string(out)
}
