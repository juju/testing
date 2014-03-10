// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
printf "%s" ` + "`basename $0`" + `
for arg in "$@"; do
  printf " \"%s\""  "$arg"
done
printf "\n"
`
)

// EnvironmentPatcher is an interface that requires just one method:
// PatchEnvironment.
type EnvironmentPatcher interface {
	PatchEnvironment(name, value string)
}

// PatchExecutable ensures that dir is in PATH and creates an executable
// in dir called execName with script as the content.
func PatchExecutable(patcher EnvironmentPatcher, dir, execName, script string) error {
	patcher.PatchEnvironment("PATH", joinPathLists(dir, os.Getenv("PATH")))
	filename := filepath.Join(dir, execName)
	return ioutil.WriteFile(filename, []byte(script), 0755)
}
