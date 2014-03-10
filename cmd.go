// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	gc "launchpad.net/gocheck"
)

// HookCommandOutput intercepts CommandOutput to a function that passes the
// actual command and it's output back via a channel, and returns the error
// passed into this function.  It also returns a cleanup function so you can
// restore the original function
func HookCommandOutput(
	outputFunc *func(cmd *exec.Cmd) ([]byte, error), output []byte, err error) (<-chan *exec.Cmd, func()) {

	cmdChan := make(chan *exec.Cmd, 1)
	origCommandOutput := *outputFunc
	cleanup := func() {
		close(cmdChan)
		*outputFunc = origCommandOutput
	}
	*outputFunc = func(cmd *exec.Cmd) ([]byte, error) {
		cmdChan <- cmd
		return output, err
	}
	return cmdChan, cleanup
}

const (
	// EchoQuotedArgs is a simple bash script that prints out the
	// basename of the command followed by the args as quoted strings.
	EchoQuotedArgs = `#!/bin/bash --norc
name=` + "`basename $0`" + `
argfile="$name.out"
rm -f $argfile
printf "%s" $name | tee -a $argfile
for arg in "$@"; do
  printf " \"%s\""  "$arg" | tee -a $argfile
done
printf "\n" | tee -a $argfile
`
)

// EnvironmentPatcher is an interface that requires just one method:
// PatchEnvironment.
type EnvironmentPatcher interface {
	PatchEnvironment(name, value string)
}

// PatchExecutable creates an executable called 'execName' in a new test
// directory and that directory is added to the path.
func PatchExecutable(c *gc.C, patcher EnvironmentPatcher, execName, script string) {
	dir := c.MkDir()
	patcher.PatchEnvironment("PATH", joinPathLists(dir, os.Getenv("PATH")))
	filename := filepath.Join(dir, execName)
	err := ioutil.WriteFile(filename, []byte(script), 0755)
	c.Assert(err, gc.IsNil)
}

// AssertEchoArgs is used to check the args from an execution of a command
// that has been patchec using PatchExecutable containing EchoQuotedArgs.
func AssertEchoArgs(c *gc.C, execName string, args ...string) {
	content, err := ioutil.ReadFile(execName + ".out")
	c.Assert(err, gc.IsNil)
	expected := execName
	for _, arg := range args {
		expected = fmt.Sprintf("%s %q", expected, arg)
	}
	expected += "\n"
	c.Assert(string(content), gc.Equals, expected)
}
