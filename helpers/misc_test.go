// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package helpers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestMiscSuite struct {
	suite.Suite
	*require.Assertions
	map1 map[string]any
	map2 map[string]any
}

type TestMiscStruct struct {
	Field1 string
	Field2 int
	field3 string
}

func (suite *TestMiscSuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())
	suite.map1 = map[string]any{
		"hello":  "world",
		"unique": "value",
		"nested": map[string]any{
			"values": "kitteh",
			"unique": "value",
		},
	}
	suite.map2 = map[string]any{
		"hello":     "it's me",
		"different": "value",
		"nested": map[string]any{
			"values":    "doggo",
			"different": "value",
		},
	}
}

func TestRetry(t *testing.T) {
	t.Run("RetriesWhenThereAreFailures", func(t *testing.T) {
		count := 0
		logCount := 0
		returnedErr := errors.New("always fail")
		countFn := func() error {
			count++
			return returnedErr
		}
		loggerFn := func(_ string, _ ...any) {
			logCount++
		}

		err := Retry(countFn, 3, 0, loggerFn)
		require.ErrorIs(t, err, returnedErr)
		require.Equal(t, 3, count)
		require.Equal(t, 5, logCount)
	})

	t.Run("ContextCancellationBeforeStart", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		fn := func() error {
			return errors.New("Never here since context got cancelled")
		}
		logger := func(_ string, _ ...any) {}
		waitThatsNotCalled := 1000000 * time.Minute
		err := RetryWithContext(ctx, fn, 5, waitThatsNotCalled, logger)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("ContextCancellationDuringExecution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		count := 0
		fn := func() error {
			count++
			if count < 2 {
				return errors.New("fail")
			}
			cancel()
			return errors.New("don't care about this error since we've cancelled and there is still another retry")
		}

		logger := func(_ string, _ ...any) {}
		// Need a teeny tiny delay here. With 0 delay the select loop may not have time to exit before the function is called again
		err := RetryWithContext(ctx, fn, 3, 5*time.Millisecond, logger)
		require.Equal(t, 2, count)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("NoErr", func(t *testing.T) {
		count := 0
		fn := func() error {
			count++
			return nil
		}

		logger := func(_ string, _ ...any) {}

		err := RetryWithContext(context.TODO(), fn, 3, 0, logger)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("InvalidAttempts", func(t *testing.T) {
		count := 0
		fn := func() error {
			count++
			return nil
		}

		logger := func(_ string, _ ...any) {}

		err := RetryWithContext(context.TODO(), fn, 0, 0, logger)
		require.Error(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("ContextCancellationDeadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
		defer cancel()

		count := 0
		fn := func() error {
			count++
			return errors.New("Always fail")
		}

		logger := func(_ string, _ ...any) {}

		err := RetryWithContext(ctx, fn, 3, 1*time.Second, logger)
		// fn should be called twice, it will wait one second after the first attempt
		// and tries to wait two seconds after the second attempt
		// but the context will cancel before the third attempt is called
		require.Equal(t, 2, count)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func (suite *TestMiscSuite) TestTransformMapKeys() {
	expected := map[string]any{
		"HELLO":  "world",
		"UNIQUE": "value",
		"NESTED": map[string]any{
			"values": "kitteh",
			"unique": "value",
		},
	}

	result := TransformMapKeys(suite.map1, strings.ToUpper)
	suite.Equal(expected, result)
}

func (suite *TestMiscSuite) TestTransformAndMergeMap() {
	expected := map[string]any{
		"DIFFERENT": "value",
		"HELLO":     "it's me",
		"UNIQUE":    "value",
		"NESTED": map[string]any{
			"values":    "doggo",
			"different": "value",
		},
	}

	result := TransformAndMergeMap(suite.map1, suite.map2, strings.ToUpper)
	suite.Equal(expected, result)
}

func (suite *TestMiscSuite) TestMergeMapRecursive() {
	expected := map[string]any{
		"different": "value",
		"hello":     "it's me",
		"unique":    "value",
		"nested": map[string]any{
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
	suite.False(result)
	result = IsNotZeroAndNotEqual(zero, original)
	suite.False(result)
	result = IsNotZeroAndNotEqual(equal, original)
	suite.False(result)
	result = IsNotZeroAndNotEqual(notEqual, original)
	suite.True(result)
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
	suite.True(*BoolPtr(true))
	suite.False(*BoolPtr(false))
	a := BoolPtr(true)
	b := BoolPtr(true)
	// This is a pointer comparison, not a value comparison
	suite.NotSame(a, b)
	suite.Equal(*a, *b)
}

func TestMisc(t *testing.T) {
	suite.Run(t, new(TestMiscSuite))
}

func (suite *TestMiscSuite) TestMergePathAndValueIntoMap() {
	type args struct {
		m     map[string]any
		path  string
		value any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    map[string]any
	}{
		{
			name:    "nested map creation",
			args:    args{m: make(map[string]any), path: "a.b.c", value: "hello"},
			wantErr: false,
			want: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "hello",
					},
				},
			},
		},
		{
			name: "overwrite existing value",
			args: args{m: map[string]any{"a": map[string]any{"b": map[string]any{"c": "initial"}}},
				path: "a.b.c", value: "updated"},
			wantErr: false,
			want: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "updated",
					},
				},
			},
		},
		{
			name:    "deeply nested map creation",
			args:    args{m: make(map[string]any), path: "a.b.c.d.e.f", value: 42},
			wantErr: false,
			want: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"d": map[string]any{
								"e": map[string]any{
									"f": 42,
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "empty path",
			args:    args{m: make(map[string]any), path: "", value: "root level"},
			wantErr: false,
			want: map[string]any{
				"": "root level",
			},
		},
		{
			name:    "root level value",
			args:    args{m: make(map[string]any), path: "root", value: "root value"},
			wantErr: false,
			want: map[string]any{
				"root": "root value",
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := MergePathAndValueIntoMap(tt.args.m, tt.args.path, tt.args.value)
			if tt.wantErr {
				suite.Error(err, "Expected an error")
			} else {
				suite.NoError(err, "Expected no error")
			}

			suite.True(reflect.DeepEqual(tt.args.m, tt.want), "Map structure mismatch: got %v, want %v", tt.args.m, tt.want)
		})
	}
}
