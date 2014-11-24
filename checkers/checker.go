// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v1"
)

func TimeBetween(start, end time.Time) gc.Checker {
	if end.Before(start) {
		return &timeBetweenChecker{end, start}
	}
	return &timeBetweenChecker{start, end}
}

type timeBetweenChecker struct {
	start, end time.Time
}

func (checker *timeBetweenChecker) Info() *gc.CheckerInfo {
	info := gc.CheckerInfo{
		Name:   "TimeBetween",
		Params: []string{"obtained"},
	}
	return &info
}

func (checker *timeBetweenChecker) Check(params []interface{}, names []string) (result bool, error string) {
	when, ok := params[0].(time.Time)
	if !ok {
		return false, "obtained value type must be time.Time"
	}
	if when.Before(checker.start) {
		return false, fmt.Sprintf("obtained time %q is before start time %q", when, checker.start)
	}
	if when.After(checker.end) {
		return false, fmt.Sprintf("obtained time %q is after end time %q", when, checker.end)
	}
	return true, ""
}

// DurationLessThan checker

type durationLessThanChecker struct {
	*gc.CheckerInfo
}

var DurationLessThan gc.Checker = &durationLessThanChecker{
	&gc.CheckerInfo{Name: "DurationLessThan", Params: []string{"obtained", "expected"}},
}

func (checker *durationLessThanChecker) Check(params []interface{}, names []string) (result bool, error string) {
	obtained, ok := params[0].(time.Duration)
	if !ok {
		return false, "obtained value type must be time.Duration"
	}
	expected, ok := params[1].(time.Duration)
	if !ok {
		return false, "expected value type must be time.Duration"
	}
	return obtained.Nanoseconds() < expected.Nanoseconds(), ""
}

// HasPrefix checker for checking strings

func stringOrStringer(value interface{}) (string, bool) {
	result, isString := value.(string)
	if !isString {
		if stringer, isStringer := value.(fmt.Stringer); isStringer {
			result, isString = stringer.String(), true
		}
	}
	return result, isString
}

type hasPrefixChecker struct {
	*gc.CheckerInfo
}

var HasPrefix gc.Checker = &hasPrefixChecker{
	&gc.CheckerInfo{Name: "HasPrefix", Params: []string{"obtained", "expected"}},
}

func (checker *hasPrefixChecker) Check(params []interface{}, names []string) (result bool, error string) {
	expected, ok := params[1].(string)
	if !ok {
		return false, "expected must be a string"
	}

	obtained, isString := stringOrStringer(params[0])
	if isString {
		return strings.HasPrefix(obtained, expected), ""
	}

	return false, "Obtained value is not a string and has no .String()"
}

type hasSuffixChecker struct {
	*gc.CheckerInfo
}

var HasSuffix gc.Checker = &hasSuffixChecker{
	&gc.CheckerInfo{Name: "HasSuffix", Params: []string{"obtained", "expected"}},
}

func (checker *hasSuffixChecker) Check(params []interface{}, names []string) (result bool, error string) {
	expected, ok := params[1].(string)
	if !ok {
		return false, "expected must be a string"
	}

	obtained, isString := stringOrStringer(params[0])
	if isString {
		return strings.HasSuffix(obtained, expected), ""
	}

	return false, "Obtained value is not a string and has no .String()"
}

type containsChecker struct {
	*gc.CheckerInfo
}

var Contains gc.Checker = &containsChecker{
	&gc.CheckerInfo{Name: "Contains", Params: []string{"obtained", "expected"}},
}

func (checker *containsChecker) Check(params []interface{}, names []string) (result bool, error string) {
	expected, ok := params[1].(string)
	if !ok {
		return false, "expected must be a string"
	}

	obtained, isString := stringOrStringer(params[0])
	if isString {
		return strings.Contains(obtained, expected), ""
	}

	return false, "Obtained value is not a string and has no .String()"
}

type sameContents struct {
	*gc.CheckerInfo
}

// SameContents checks that the obtained slice contains all the values (and
// same number of values) of the expected slice and vice versa, without respect
// to order or duplicates. Uses DeepEquals on mapped contents to compare.
var SameContents gc.Checker = &sameContents{
	&gc.CheckerInfo{Name: "SameContents", Params: []string{"obtained", "expected"}},
}

func (checker *sameContents) Check(params []interface{}, names []string) (result bool, error string) {
	if len(params) != 2 {
		return false, "SameContents expects two slice arguments"
	}
	obtained := params[0]
	expected := params[1]

	tob := reflect.TypeOf(obtained)
	if tob.Kind() != reflect.Slice {
		return false, fmt.Sprintf("SameContents expects the obtained value to be a slice, got %q",
			tob.Kind())
	}

	texp := reflect.TypeOf(expected)
	if texp.Kind() != reflect.Slice {
		return false, fmt.Sprintf("SameContents expects the expected value to be a slice, got %q",
			texp.Kind())
	}

	if texp != tob {
		return false, fmt.Sprintf(
			"SameContents expects two slices of the same type, expected: %q, got: %q",
			texp, tob)
	}

	vexp := reflect.ValueOf(expected)
	vob := reflect.ValueOf(obtained)
	length := vexp.Len()

	if vob.Len() != length {
		// Slice has incorrect number of elements
		return false, ""
	}

	// spin up maps with the entries as keys and the counts as values
	mob := make(map[interface{}]int, length)
	mexp := make(map[interface{}]int, length)

	for i := 0; i < length; i++ {
		mexp[vexp.Index(i).Interface()]++
		mob[vob.Index(i).Interface()]++
	}
	return reflect.DeepEqual(mob, mexp), ""
}

type isNilChecker struct {
	*gc.CheckerInfo
}

// The IsNil checker tests whether the obtained value is nil.
// Copied from the gocheck package with the addition of printing out
// the error stack if the value is an error and supports the ErrorStack
// method.
//
// For example:
//
//    c.Assert(err, IsNil)
//
var IsNil gc.Checker = &isNilChecker{
	&gc.CheckerInfo{Name: "IsNil", Params: []string{"value"}},
}

type ErrorStacker interface {
	error
	StackTrace() []string
}

func (checker *isNilChecker) Check(params []interface{}, names []string) (bool, string) {
	result := isNil(params[0])
	message := ""
	if !result {
		if stacker, ok := params[0].(ErrorStacker); ok {
			stack := stacker.StackTrace()
			if stack != nil {
				message = "error stack:\n\t" + strings.Join(stack, "\n\t")
			}
		}
	}
	return result, message
}

func isNil(obtained interface{}) (result bool) {
	if obtained == nil {
		result = true
	} else {
		switch v := reflect.ValueOf(obtained); v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			return v.IsNil()
		}
	}
	return
}

type codecEqualChecker struct {
	name      string
	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error
}

// JSONEquals defines a checker that checks whether a byte slice, when
// unmarshaled as JSON, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in JSON and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
var JSONEquals = &codecEqualChecker{
	name:      "JSONEquals",
	marshal:   json.Marshal,
	unmarshal: json.Unmarshal,
}

// YAMLEquals defines a checker that checks whether a byte slice, when
// unmarshaled as YAML, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in YAML and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
var YAMLEquals = &codecEqualChecker{
	name:      "YAMLEquals",
	marshal:   yaml.Marshal,
	unmarshal: yaml.Unmarshal,
}

func (checker *codecEqualChecker) Info() *gc.CheckerInfo {
	return &gc.CheckerInfo{
		Name:   checker.name,
		Params: []string{"obtained", "expected"},
	}
}

func (checker *codecEqualChecker) Check(params []interface{}, names []string) (result bool, error string) {
	gotContent, ok := params[0].([]byte)
	if !ok {
		return false, fmt.Sprintf("expected []byte, got %T", params[0])
	}
	expectContent := params[1]
	expectContentBytes, err := checker.marshal(expectContent)
	if err != nil {
		return false, fmt.Sprintf("cannot marshal expected contents: %v", err)
	}
	var expectContentVal interface{}
	if err := checker.unmarshal(expectContentBytes, &expectContentVal); err != nil {
		return false, fmt.Sprintf("cannot unmarshal expected contents: %v", err)
	}

	var gotContentVal interface{}
	if err := checker.unmarshal(gotContent, &gotContentVal); err != nil {
		return false, fmt.Sprintf("cannot unmarshal obtained contents: %v; %q", err, gotContent)
	}

	if ok, err := jc.DeepEqual(gotContentVal, expectContentVal); !ok {
		return false, err.Error()
	}
	return true, ""
}
