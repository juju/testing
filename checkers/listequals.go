// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers

import (
	"fmt"
	"reflect"

	gc "gopkg.in/check.v1"
)

type listEqualsChecker struct {
	*gc.CheckerInfo
}

// The ListEquals checker verifies if two lists are equal. If they are not,
// it will essentially run a "diff" algorithm to provide the developer with
// an easily understandable summary of the difference between the two lists.
var ListEquals gc.Checker = &listEqualsChecker{
	&gc.CheckerInfo{Name: "ListEquals", Params: []string{"obtained", "expected"}},
}

func (l *listEqualsChecker) Check(params []interface{}, names []string) (result bool, error string) {
	obtained := params[0]
	expected := params[1]

	// Do some simple pre-checks. First, that both 'obtained' and 'expected'
	// are indeed slices.
	vExp := reflect.ValueOf(expected)
	if vExp.Kind() != reflect.Slice {
		return false, fmt.Sprintf("expected value is not a slice")
	}

	vObt := reflect.ValueOf(obtained)
	if vObt.Kind() != reflect.Slice {
		return false, fmt.Sprintf("obtained value is not a slice")
	}

	// Check that element types are the same
	expElemType := vExp.Type().Elem()
	obtElemType := vObt.Type().Elem()

	if expElemType != obtElemType {
		return false, fmt.Sprintf("element types are not equal")
	}

	// Check that the element type is comparable.
	if !expElemType.Comparable() {
		return false, fmt.Sprintf("element type is not comparable")
	}

	// The approach here is to find a longest-common subsequence using dynamic
	// programming, and use this to generate the diff. This algorithm runs in
	// O(n^2). However, naive list equality is only O(n). Hence, to be more
	// efficient, we should first check if the lists are equal, and if they are
	// not, we do the more complicated work to find out exactly *how* they are
	// different.

	slicesEqual := true
	// Check length is equal
	if vObt.Len() == vExp.Len() {
		// Iterate through and check every element
		for i := 0; i < vExp.Len(); i++ {
			a := vObt.Index(i)
			b := vExp.Index(i)
			if !a.Equal(b) {
				slicesEqual = false
				break
			}
		}

		if slicesEqual {
			return true, ""
		}
	}

	// If we're here, the lists are not equal, so run the DP algorithm to
	// compute the diff.
	return false, generateDiff(vObt, vExp)
}

func generateDiff(obtained, expected reflect.Value) string {
	// lenLCS[m][n] stores the length of the longest common subsequence of
	// obtained[:m] and expected[:n]
	lenLCS := make([][]int, obtained.Len()+1)
	for i := 0; i <= obtained.Len(); i++ {
		lenLCS[i] = make([]int, expected.Len()+1)
	}

	// lenLCS[i][0] and lenLCS[0][j] are already correctly initialised to 0

	for i := 1; i <= obtained.Len(); i++ {
		for j := 1; j <= expected.Len(); j++ {
			if obtained.Index(i - 1).Equal(expected.Index(j - 1)) {
				// We can extend the longest subsequence of obtained[:i-1] and expected[:j-1]
				lenLCS[i][j] = lenLCS[i-1][j-1] + 1
			} else {
				// We can't extend a previous subsequence
				lenLCS[i][j] = max(lenLCS[i-1][j], lenLCS[i][j-1])
			}
		}
	}

	// "Traceback" to calculate the diff
	var diffs []diff
	i := obtained.Len()
	j := expected.Len()

	for i > 0 && j > 0 {
		if lenLCS[i][j] == lenLCS[i-1][j-1] {
			// Element changed at this index
			diffs = append(diffs, elementChanged{j - 1, expected.Index(j - 1), obtained.Index(i - 1)})
			i -= 1
			j -= 1

		} else if lenLCS[i][j] == lenLCS[i-1][j] {
			// Additional/unexpected element at this index
			diffs = append(diffs, elementAdded{j, obtained.Index(i - 1)})
			i -= 1

		} else if lenLCS[i][j] == lenLCS[i][j-1] {
			// Element missing at this index
			diffs = append(diffs, elementRemoved{j - 1, expected.Index(j - 1)})
			j -= 1

		} else {
			// Elements are the same at this index - no diff
			i -= 1
			j -= 1
		}
	}
	for i > 0 {
		// Extra elements have been added at the start
		diffs = append(diffs, elementAdded{0, obtained.Index(i - 1)})
		i -= 1
	}
	for j > 0 {
		// Elements are missing at the start
		diffs = append(diffs, elementRemoved{j - 1, expected.Index(j - 1)})
		j -= 1
	}

	// Convert diffs array into human-readable error
	description := "difference:"
	for k := len(diffs) - 1; k >= 0; k-- {
		description += "\n    - " + diffs[k].String()
	}
	return description
}

// diff represents a single difference between the two slices.
type diff interface {
	// Prints a user friendly description of this difference.
	String() string
}

type elementAdded struct {
	index   int
	element any
}

func (d elementAdded) String() string {
	return fmt.Sprintf("at index %d: unexpected element %#v", d.index, d.element)
}

type elementChanged struct {
	index int

	original, changed any
}

func (d elementChanged) String() string {
	return fmt.Sprintf("at index %d: obtained element %#v, expected %#v", d.index, d.changed, d.original)
}

type elementRemoved struct {
	index   int
	element any
}

func (d elementRemoved) String() string {
	return fmt.Sprintf("at index %d: missing element %#v", d.index, d.element)
}
