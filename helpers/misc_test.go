// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package helpers

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestMiscSuite struct {
	suite.Suite
	*require.Assertions
	map1 map[string]interface{}
	map2 map[string]interface{}
}

type TestMiscStruct struct {
	Field1 string
	Field2 int
	field3 string
}

func (suite *TestMiscSuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())
	suite.map1 = map[string]interface{}{
		"hello":  "world",
		"unique": "value",
		"nested": map[string]interface{}{
			"values": "kitteh",
			"unique": "value",
		},
	}
	suite.map2 = map[string]interface{}{
		"hello":     "it's me",
		"different": "value",
		"nested": map[string]interface{}{
			"values":    "doggo",
			"different": "value",
		},
	}
}

func (suite *TestMiscSuite) TestRetry() {
	var count int
	countFn := func() error {
		count++
		if count < 4 {
			return errors.New("count exceeded")
		}
		return nil
	}
	var logCount int
	loggerFn := func(_ string, _ ...any) {
		logCount++
	}

	count = 0
	logCount = 0
	err := Retry(countFn, 3, 0, loggerFn)
	suite.Error(err)
	suite.Equal(3, count)
	suite.Equal(3, logCount)

	count = 0
	logCount = 0
	err = Retry(countFn, 4, 0, loggerFn)
	suite.NoError(err)
	suite.Equal(4, count)
	suite.Equal(3, logCount)
}

func (suite *TestMiscSuite) TestMergeMap() {
	expected := map[string]interface{}{
		"different": "value",
		"hello":     "it's me",
		"unique":    "value",
		"nested": map[string]interface{}{
			"values":    "doggo",
			"different": "value",
		},
	}

	result := MergeMap(suite.map1, suite.map2)
	suite.Equal(expected, result)
}

func (suite *TestMiscSuite) TestTransformMapKeys() {
	expected := map[string]interface{}{
		"HELLO":  "world",
		"UNIQUE": "value",
		"NESTED": map[string]interface{}{
			"values": "kitteh",
			"unique": "value",
		},
	}

	result := TransformMapKeys(suite.map1, strings.ToUpper)
	suite.Equal(expected, result)
}

func (suite *TestMiscSuite) TestTransformAndMergeMap() {
	expected := map[string]interface{}{
		"DIFFERENT": "value",
		"HELLO":     "it's me",
		"UNIQUE":    "value",
		"NESTED": map[string]interface{}{
			"values":    "doggo",
			"different": "value",
		},
	}

	result := TransformAndMergeMap(suite.map1, suite.map2, strings.ToUpper)
	suite.Equal(expected, result)
}

func (suite *TestMiscSuite) TestMergeMapRecursive() {
	expected := map[string]interface{}{
		"different": "value",
		"hello":     "it's me",
		"unique":    "value",
		"nested": map[string]interface{}{
			"values":    "doggo",
			"different": "value",
			"unique":    "value",
		},
	}

	result := MergeMapRecursive(suite.map1, suite.map2)
	suite.Equal(expected, result)
}

func (suite *TestMiscSuite) TestIsNotZeroAndNotEqual() {
	original := TestMiscStruct{
		Field1: "hello",
		Field2: 100,
		field3: "world",
	}
	zero := TestMiscStruct{}
	equal := TestMiscStruct{
		Field1: "hello",
	}
	notEqual := TestMiscStruct{
		Field1: "kitteh",
	}

	result := IsNotZeroAndNotEqual(original, original)
	suite.Equal(false, result)
	result = IsNotZeroAndNotEqual(zero, original)
	suite.Equal(false, result)
	result = IsNotZeroAndNotEqual(equal, original)
	suite.Equal(false, result)
	result = IsNotZeroAndNotEqual(notEqual, original)
	suite.Equal(true, result)
}

func (suite *TestMiscSuite) TestMergeNonZero() {
	original := TestMiscStruct{
		Field1: "hello",
		Field2: 100,
		field3: "world",
	}
	overrides := TestMiscStruct{
		Field1: "kitteh",
		Field2: 300,
		// field 3 is private and shouldn't be set (but also shouldn't panic)
		field3: "doggo",
	}

	result := MergeNonZero(original, overrides)
	suite.Equal("kitteh", result.Field1)
	suite.Equal(300, result.Field2)
	suite.Equal("world", result.field3)

	withZero := TestMiscStruct{
		Field1: "kitteh",
	}

	result = MergeNonZero(original, withZero)
	suite.Equal("kitteh", result.Field1)
	suite.Equal(100, result.Field2)
	suite.Equal("world", result.field3)
}

func (suite *TestMiscSuite) TestBoolPtr() {
	suite.Equal(true, *BoolPtr(true))
	suite.Equal(false, *BoolPtr(false))
	a := BoolPtr(true)
	b := BoolPtr(true)
	// This is a pointer comparison, not a value comparison
	suite.False(a == b)
	suite.True(*a == *b)
}

func TestMisc(t *testing.T) {
	suite.Run(t, new(TestMiscSuite))
}
