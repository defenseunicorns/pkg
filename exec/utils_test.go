// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package exec

import (
	"errors"
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
		{cmd: "./zarf version", pref: ShellPreference{}, os: "linux", want: "zarf version"},
	}

	// Run tests without registering command mutations
	for _, tc := range tests {
		got := mutateCommandForOS(tc.cmd, tc.pref, tc.os)
		if got != tc.want {
			t.Fatalf("wanted: %s, got: %s", tc.want, got)
		}
	}

	// Dogsled the error here since we are not explicitly testing this here
	_ = RegisterCmdMutation("zarf", "/usr/local/bin/zarf")

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

func TestRegisterCmdMutation(t *testing.T) {
	type test struct {
		cmdKey  string
		cmdLoc  string
		wantLoc string
		wantOk  bool
		wantErr error
	}

	tests := []test{
		{cmdKey: "uds", cmdLoc: "/usr/local/bin/uds", wantLoc: "", wantOk: false, wantErr: nil},
		{cmdKey: "uds", cmdLoc: "/usr/local/bin/uds", wantLoc: "/usr/local/bin/uds", wantOk: true, wantErr: nil},
		{cmdKey: "kubectl", cmdLoc: "/usr/local/bin/kubectl", wantLoc: "", wantOk: false, wantErr: nil},
		{cmdKey: "kubectl", cmdLoc: "/usr/local/bin/kubectl", wantLoc: "/usr/local/bin/kubectl", wantOk: true, wantErr: nil},
		{cmdKey: "kitteh", cmdLoc: "/usr/local/bin/kitteh", wantLoc: "", wantOk: false, wantErr: errors.New("kitteh is not a supported command key")},
	}

	for _, tc := range tests {
		gotLoc, gotOk := GetCmdMutation(tc.cmdKey)
		if gotOk != tc.wantOk {
			t.Fatalf("wanted: %t, got: %t", tc.wantOk, gotOk)
		}
		if gotLoc != tc.wantLoc {
			t.Fatalf("wanted: %s, got: %s", tc.wantLoc, gotLoc)
		}

		gotErr := RegisterCmdMutation(tc.cmdKey, tc.cmdLoc)
		if gotErr != nil && tc.wantErr != nil {
			if gotErr.Error() != tc.wantErr.Error() {
				t.Fatalf("wanted err: %s, got err: %s", tc.wantErr, gotErr)
			}
		} else if gotErr != nil {
			t.Fatalf("got unexpected err: %s", gotErr)
		}
	}
}
