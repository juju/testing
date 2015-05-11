// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	gc "gopkg.in/check.v1"
)

// These are generally useful tags to use in tests.
const (
	TagBase       = "base"
	TagSmall      = "small"
	TagMedium     = "medium"
	TagLarge      = "large"
	TagSmoke      = "smoke"      // Fast sanity-check.
	TagFunctional = "functional" // Does not use test doubles for low-level.
	// TODO(ericsnow) Add other tags? For example:
	//  - external: test uses external resources (e.g. filesystem)
	//  - cloud: test interacts with a cloud provider's API
	//  - vm: test runs in a local VM (e.g. kvm) for isolation
)

var defaultTags = []string{
	TagBase,
}

var (
	includedTags []string
	excludedTags []string
)

func init() {
	smoke := flag.Bool("smoke", false, "Run the basic set of fast tests.")
	include := flag.String("include-tags", "", "Tagged tests to run.")
	exclude := flag.String("exclude-tags", "", "Tagged tests to not run.")
	flag.Parse()

	includedTags = parseTags(*include)
	if *smoke {
		includedTags = append(includedTags, TagSmoke)
	}
	if len(includedTags) == 0 {
		includedTags = defaultTags
	}
	// TODO(ericsnow) support implied tags (e.g. VM -> Large)?

	excludedTags = parseTags(*exclude)
}

func parseTags(rawList ...string) []string {
	var tags []string
	for _, raw := range rawList {
		for _, entry := range strings.Split(raw, ",") {
			if len(entry) == 0 {
				continue
			}
			tag := entry
			tags = append(tags, tag)
		}
	}
	return tags
}

// CheckTag determines whether or not any of the given tags were passed
// in at the commandline.
func CheckTag(tags ...string) bool {
	return MatchTag(tags...) != ""
}

// MatchTag returns the first provided tag that matches the ones passed
// in at the commandline.
func MatchTag(tags ...string) string {
	for _, tag := range tags {
		for _, excludedTag := range excludedTags {
			if tag == excludedTag {
				return ""
			}
		}
	}

	for _, tag := range tags {
		for _, includedTag := range includedTags {
			if tag == includedTag {
				return tag
			}
		}
	}
	return ""
}

// RegisterPackageTagged registers the package for testing if any of the
// given tags were passed in at the commandline.
func RegisterPackageTagged(t *testing.T, tags ...string) {
	if CheckTag(tags...) {
		gc.TestingT(t)
	}
}

// SuiteTagged registers the suite with the test runner if any of the
// given tags were passed in at the commandline.
func SuiteTagged(suite interface{}, tags ...string) {
	if CheckTag(tags...) {
		gc.Suite(suite)
	}
}

// RequireTag causes a test or suite to skip if none of the given tags
// were passed in at the commandline.
func RequireTag(c *gc.C, tags ...string) {
	if !CheckTag(tags...) {
		c.Skip(fmt.Sprintf("skipping due to no matching tags (%v)", tags))
	}
}

// SkipTag causes a test or suite to skip if any of the given tags were
// passed in at the commandline.
func SkipTag(c *gc.C, tags ...string) {
	matched := MatchTag(tags...)
	if matched != "" {
		c.Skip(fmt.Sprintf("skipping due to %q tag", matched))
	}
}
