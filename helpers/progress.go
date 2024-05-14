// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package helpers

import "io"

// Write doesn't do anything but satisfy implementation
func (DiscardProgressWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Close doesn't do anything but satisfy implementation
func (DiscardProgressWriter) Close() {}

// Updatef doesn't do anything but satisfy implementation
func (DiscardProgressWriter) Updatef(_ string, _ ...any) {}

// Successf doesn't do anything but satisfy implementation
func (DiscardProgressWriter) Successf(_ string, _ ...any) {}

// Errorf doesn't do anything but satisfy implementation
func (DiscardProgressWriter) Errorf(_ string, _ ...any) {}

// DiscardProgressWriter is a ProgressWriter in which all calls succeed without doing anything
// Use this or nil or if you don't care about writing progress
type DiscardProgressWriter struct{}

// ProgressWriter wraps io.Writer, but also includes functions to give the user
// additional context on what's going on. Useful in OCI for tracking layers
type ProgressWriter interface {
	Updatef(string, ...any)
	Successf(string, ...any)
	Errorf(string, ...any)
	io.Writer
	io.Closer
}
