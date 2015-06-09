// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

var (
	HandleCommandline = handleCommandline
	ParseTags         = parseTags
)

func GetTags() [][]string {
	return rawTags.parse()
}

func SetTags(tags ...[]string) {
	rawTags.parsed = tags
}
