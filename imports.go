// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing

import (
	"fmt"
	"go/build"
	"sort"
	"strings"
)

// FindImports returns a sorted list of packages imported by the package
// with the given name that have the given prefix. The resulting list
// removes the common prefix, leaving just the short names.
func FindImports(packageName, prefix string) ([]string, error) {
	allPkgs := make(map[string]bool)
	if err := findImports(packageName, allPkgs); err != nil {
		return nil, err
	}
	var result []string
	for name := range allPkgs {
		if strings.HasPrefix(name, prefix) {
			result = append(result, name[len(prefix):])
		}
	}
	sort.Strings(result)
	return result, nil
}

// findImports recursively adds all imported packages of given
// package (packageName) to allPkgs map.
func findImports(packageName string, allPkgs map[string]bool) error {
	pkg, err := build.Default.Import(packageName, "", 0)
	if err != nil {
		return fmt.Errorf("cannot find %q: %v", packageName, err)
	}
	for _, name := range pkg.Imports {
		if !allPkgs[name] {
			allPkgs[name] = true
			if err := findImports(name, allPkgs); err != nil {
				return err
			}
		}
	}
	return nil
}
