// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers

import (
	"fmt"
	"regexp"

	gc "gopkg.in/check.v1"
)

// MultiChecker is a deep checker that by default matches for equality.
// But checks can be overriden based on path (either explicit match or regexp)
type MultiChecker struct {
	*gc.CheckerInfo
	checks      map[string]multiCheck
	regexChecks []regexCheck
}

type multiCheck struct {
	checker gc.Checker
	args    []interface{}
}

type regexCheck struct {
	multiCheck
	regex *regexp.Regexp
}

// NewMultiChecker creates a MultiChecker which is a deep checker that by default matches for equality.
// But checks can be overriden based on path (either explicit match or regexp)
func NewMultiChecker() *MultiChecker {
	return &MultiChecker{
		CheckerInfo: &gc.CheckerInfo{Name: "MultiChecker", Params: []string{"obtained", "expected"}},
		checks:      make(map[string]multiCheck),
	}
}

// Add an explict checker by path.
func (checker *MultiChecker) Add(path string, c gc.Checker, args ...interface{}) *MultiChecker {
	checker.checks[path] = multiCheck{
		checker: c,
		args:    args,
	}
	return checker
}

// AddRegex exception which matches path with regex.
func (checker *MultiChecker) AddRegex(pathRegex string, c gc.Checker, args ...interface{}) *MultiChecker {
	checker.regexChecks = append(checker.regexChecks, regexCheck{
		multiCheck: multiCheck{
			checker: c,
			args:    args,
		},
		regex: regexp.MustCompile("^" + pathRegex + "$"),
	})
	return checker
}

// Check for go check Checker interface.
func (checker *MultiChecker) Check(params []interface{}, names []string) (result bool, errStr string) {
	customCheckFunc := func(path string, a1 interface{}, a2 interface{}) (useDefault bool, equal bool, err error) {
		var mc *multiCheck
		if c, ok := checker.checks[path]; ok {
			mc = &c
		} else {
			for _, v := range checker.regexChecks {
				if v.regex.MatchString(path) {
					mc = &v.multiCheck
					break
				}
			}
		}
		if mc == nil {
			return true, false, nil
		}

		params := append([]interface{}{a1}, mc.args...)
		info := mc.checker.Info()
		if len(params) < len(info.Params) {
			return false, false, fmt.Errorf("Wrong number of parameters for %s: want %d, got %d", info.Name, len(info.Params), len(params)+1)
		}
		// Copy since it may be mutated by Check.
		names := append([]string{}, info.Params...)

		// Trim to the expected params len.
		params = params[:len(info.Params)]

		// Perform substitution
		for i, v := range params {
			if v == ExpectedValue {
				params[i] = a2
			}
		}

		result, errStr := mc.checker.Check(params, names)
		if result {
			return false, true, nil
		}
		if path == "" {
			path = "top level"
		}
		return false, false, fmt.Errorf("mismatch at %s: %s", path, errStr)
	}
	if ok, err := DeepEqualWithCustomCheck(params[0], params[1], customCheckFunc); !ok {
		return false, err.Error()
	}
	return true, ""
}

// ExpectedValue if passed to MultiChecker.Add or MultiChecker.AddRegex, will be substituded with the expected value.
var ExpectedValue = &struct{}{}
