// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

// +build windows

package testing

import "os"

// We cannot yet send an os.Interrupt on Windows
// For now this function is equivalent to Destroy
// https://code.google.com/p/go/source/browse/src/pkg/os/doc.go?spec=svne165495e81bfe6fbdd44ef99e9266bb7d09dae67&name=e165495e81bf&r=e165495e81bfe6fbdd44ef99e9266bb7d09dae67#49
func (inst *MgoInstance) DestroyWithLog() {
	inst.killAndCleanup(os.Kill)
}

