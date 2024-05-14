// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package exec

import (
	"testing"
)

func TestGetOSShell(t *testing.T) {
	type test struct {
		shell     Shell
		os        string
		wantShell string
		wantArgs  []string
	}

	tests := []test{
		{shell: Shell{}, os: "linux", wantShell: "sh", wantArgs: []string{"-e", "-c"}},
		{shell: Shell{Windows: "pwsh"}, os: "linux", wantShell: "sh", wantArgs: []string{"-e", "-c"}},
		{shell: Shell{Linux: "pwsh"}, os: "linux", wantShell: "pwsh", wantArgs: []string{"-Command", "$ErrorActionPreference = 'Stop';"}},
		{shell: Shell{}, os: "darwin", wantShell: "sh", wantArgs: []string{"-e", "-c"}},
		{shell: Shell{Windows: "pwsh"}, os: "darwin", wantShell: "sh", wantArgs: []string{"-e", "-c"}},
		{shell: Shell{Darwin: "pwsh"}, os: "darwin", wantShell: "pwsh", wantArgs: []string{"-Command", "$ErrorActionPreference = 'Stop';"}},
		{shell: Shell{}, os: "windows", wantShell: "powershell", wantArgs: []string{"-Command", "$ErrorActionPreference = 'Stop';"}},
		{shell: Shell{Linux: "pwsh"}, os: "windows", wantShell: "powershell", wantArgs: []string{"-Command", "$ErrorActionPreference = 'Stop';"}},
		{shell: Shell{Windows: "cmd"}, os: "windows", wantShell: "cmd", wantArgs: []string{"/c"}},
	}

	// Run tests without registering command mutations
	for _, tc := range tests {
		gotShell, gotArgs := getOSShellForOS(tc.shell, tc.os)
		if gotShell != tc.wantShell {
			t.Fatalf("wanted shell: %s, got shell: %s", tc.wantShell, gotShell)
		}

		if len(gotArgs) != len(tc.wantArgs) {
			t.Fatalf("wanted args len: %d, got args len: %d", len(tc.wantArgs), len(gotArgs))
		}

		for idx := range gotArgs {
			if gotArgs[idx] != tc.wantArgs[idx] {
				t.Fatalf("at index %d: wanted arg: %s, got arg: %s", idx, tc.wantShell, gotShell)
			}
		}
	}
}
