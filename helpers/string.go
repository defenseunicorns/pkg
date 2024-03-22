// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present The Defense Unicorn Pkg Authors

// Package helpers provides generic helper functions
package helpers

// First30last30 returns the source string that has been trimmed to 30 characters at the beginning and end.
func First30last30(s string) string {
	if len(s) > 60 {
		return s[0:27] + "..." + s[len(s)-26:]
	}

	return s
}
