// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package exec

import (
	"testing"
)

func TestMutateCommand(t *testing.T) {
	type test struct {
		cmd  string
		pref ShellPreference
		os   string
		want string
	}

	tests := []test{
		{cmd: "touch \"hello\"", pref: ShellPreference{}, os: "linux", want: "touch \"hello\""},
		{cmd: "touch \"hello\"", pref: ShellPreference{}, os: "windows", want: "New-Item \"hello\""},
		{cmd: "echo \"${hello}\"", pref: ShellPreference{}, os: "linux", want: "echo \"${hello}\""},
		{cmd: "echo \"${hello}\"", pref: ShellPreference{}, os: "windows", want: "echo \"$Env:hello\""},
		{cmd: "echo \"${hello}\"", pref: ShellPreference{Windows: "pwsh"}, os: "windows", want: "echo \"$Env:hello\""},
		{cmd: "echo \"${hello}\"", pref: ShellPreference{Windows: "cmd"}, os: "windows", want: "echo \"${hello}\""},
		{cmd: "./zarf version", pref: ShellPreference{}, os: "linux", want: "./zarf version"},
	}

	// Run tests without registering command mutations
	for _, tc := range tests {
		got := mutateCommandForOS(tc.cmd, tc.pref, tc.os)
		if got != tc.want {
			t.Fatalf("wanted: %s, got: %s", tc.want, got)
		}
	}

	RegisterCmdMutation("zarf", "/usr/local/bin/zarf")

	tests = []test{
		{cmd: "./zarf version", pref: ShellPreference{}, os: "linux", want: "/usr/local/bin/zarf version"},
	}

	// Run tests after registering command mutations
	for _, tc := range tests {
		got := mutateCommandForOS(tc.cmd, tc.pref, tc.os)
		if got != tc.want {
			t.Fatalf("wanted: %s, got: %s", tc.want, got)
		}
	}
}
