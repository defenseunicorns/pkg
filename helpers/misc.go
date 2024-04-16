// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package helpers

import (
	"context"
	"fmt"
	"maps"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// BoolPtr returns a pointer to a bool.
func BoolPtr(b bool) *bool {
	return &b
}

// RetryWithContext retries a function until it succeeds, the timeout is reached, or the context is done.
// The delay between attempts increases exponentially as (2^(attempt-1)) * delay.
// For example, with a delay of one second and three retries, the timing would be:
// - First attempt: immediate
// - Second attempt: after one second
// - Third attempt: after two seconds
func RetryWithContext(ctx context.Context, fn func() error, attempts int, delay time.Duration, logger func(format string, args ...any)) error {
	var err error
	for r := 0; r < attempts; r++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = fn()
			if err == nil {
				return nil
			}

			logger("Attempt (%d/%d) failed with: %s", r+1, attempts, err.Error())

			// No reason to wait when we aren't going to retry again
			if r+1 == attempts {
				return err
			}

			pow := math.Pow(2, float64(r))
			backoff := delay * time.Duration(pow)

			logger("Retrying in %s", backoff)

			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return ctx.Err()
			}
		}
	}

	return err
}

// Retry retries a function until it succeeds, the timeout is reached, or the context is done.
// The delay between attempts increases exponentially as (2^(attempt-1)) * delay.
// For example, with a delay of one second and three retries, the timing would be:
// - First attempt: immediate
// - Second attempt: after one second
// - Third attempt: after two seconds
func Retry(fn func() error, attempts int, delay time.Duration, logger func(format string, args ...any)) error {
	return RetryWithContext(context.Background(), fn, attempts, delay, logger)
}

// TransformMapKeys takes a map and transforms its keys using the provided function.
func TransformMapKeys[T any](m map[string]T, transform func(string) string) (r map[string]T) {
	r = map[string]T{}

	for key, value := range m {
		r[transform(key)] = value
	}

	return r
}

// TransformAndMergeMap transforms keys in both maps then merges map m2 with m1 overwriting common values with m2's values.
func TransformAndMergeMap[T any](m1, m2 map[string]T, transform func(string) string) (r map[string]T) {
	r = TransformMapKeys(m1, transform)
	mt2 := TransformMapKeys(m2, transform)
	maps.Copy(r, mt2)
	return r
}

// MergeMapRecursive recursively (nestedly) merges map m2 with m1 overwriting common values with m2's values.
func MergeMapRecursive(m1, m2 map[string]interface{}) (r map[string]interface{}) {
	r = map[string]interface{}{}

	for key, value := range m1 {
		r[key] = value
	}

	for key, value := range m2 {
		if value, ok := value.(map[string]interface{}); ok {
			if nestedValue, ok := r[key]; ok {
				if nestedValue, ok := nestedValue.(map[string]interface{}); ok {
					r[key] = MergeMapRecursive(nestedValue, value)
					continue
				}
			}
		}
		r[key] = value
	}

	return r
}

// MatchRegex wraps a get function around a substring match.
func MatchRegex(regex *regexp.Regexp, str string) (func(string) string, error) {
	// Validate the string.
	matches := regex.FindStringSubmatch(str)

	// Parse the string into its components.
	get := func(name string) string {
		return matches[regex.SubexpIndex(name)]
	}

	if len(matches) == 0 {
		return get, fmt.Errorf("unable to match against %s", str)
	}

	return get, nil
}

// IsNotZeroAndNotEqual is used to test if a struct has zero values or is equal values with another struct
func IsNotZeroAndNotEqual[T any](given T, equal T) bool {
	givenValue := reflect.ValueOf(given)
	equalValue := reflect.ValueOf(equal)

	if givenValue.NumField() != equalValue.NumField() {
		return true
	}

	for i := 0; i < givenValue.NumField(); i++ {
		if !givenValue.Field(i).IsZero() &&
			givenValue.Field(i).CanInterface() &&
			givenValue.Field(i).Interface() != equalValue.Field(i).Interface() {

			return true
		}
	}
	return false
}

// MergeNonZero is used to merge non-zero overrides from one struct into another of the same type
func MergeNonZero[T any](original T, overrides T) T {
	originalValue := reflect.ValueOf(&original)
	overridesValue := reflect.ValueOf(&overrides)

	for i := 0; i < originalValue.Elem().NumField(); i++ {
		if !overridesValue.Elem().Field(i).IsZero() &&
			overridesValue.Elem().Field(i).CanSet() {

			overrideField := overridesValue.Elem().Field(i)
			originalValue.Elem().Field(i).Set(overrideField)
		}
	}
	return originalValue.Elem().Interface().(T)
}

// MergePathAndValueIntoMap takes a path in dot notation as a string and a value (also as a string for simplicity),
// then merges this into the provided map. The value can be any type.
func MergePathAndValueIntoMap(m map[string]any, path string, value any) error {
	pathParts := strings.Split(path, ".")
	current := m
	for i, part := range pathParts {
		if i == len(pathParts)-1 {
			// Set the value at the last key in the path.
			current[part] = value
		} else {
			if _, exists := current[part]; !exists {
				// If the part does not exist, create a new map for it.
				current[part] = make(map[string]any)
			}

			nextMap, ok := current[part].(map[string]any)
			if !ok {
				return fmt.Errorf("conflict at %q, expected map but got %T", strings.Join(pathParts[:i+1], "."), current[part])
			}
			current = nextMap
		}
	}
	return nil
}
