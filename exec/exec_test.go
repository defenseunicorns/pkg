// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package exec

import (
	"bytes"
	"errors"
	"testing"
)

func TestCmd(t *testing.T) {
	type test struct {
		config     Config
		command    string
		args       []string
		wantStdOut string
		wantStdErr string
		wantErr    error
	}

	var stdOutBuff bytes.Buffer
	var stdErrBuff bytes.Buffer

	tests := []test{
		{wantErr: errors.New("command is required")},
		{config: Config{}, command: "echo", args: []string{"hello kitteh"}, wantStdOut: "hello kitteh\n"},
		{config: Config{Env: []string{"ARCH=amd64"}}, command: "printenv", args: []string{"ARCH"}, wantStdOut: "amd64\n"},
		{config: Config{Dir: "/"}, command: "pwd", wantStdOut: "/\n"},
		{config: Config{Stdout: &stdOutBuff}, command: "sh", args: []string{"-c", "echo \"hello kitteh out\""}, wantStdOut: "hello kitteh out\n"},
		{config: Config{Stderr: &stdErrBuff}, command: "sh", args: []string{"-c", "echo \"hello kitteh err\" >&2"}, wantStdErr: "hello kitteh err\n"},
	}

	// Run tests without registering command mutations
	for _, tc := range tests {
		gotStdOut, gotStdErr, gotErr := Cmd(tc.config, tc.command, tc.args...)
		if gotStdOut != tc.wantStdOut {
			t.Fatalf("wanted std out: %s, got std out: %s", tc.wantStdOut, gotStdOut)
		}
		if gotStdErr != tc.wantStdErr {
			t.Fatalf("wanted std err: %s, got std err: %s", tc.wantStdErr, gotStdErr)
		}
		if gotErr != nil && tc.wantErr != nil {
			if gotErr.Error() != tc.wantErr.Error() {
				t.Fatalf("wanted err: %s, got err: %s", tc.wantErr, gotErr)
			}
		} else if gotErr != nil {
			t.Fatalf("got unexpected err: %s", gotErr)
		}
	}

	stdOutBufferString := stdOutBuff.String()
	if stdOutBufferString != "hello kitteh out\n" {
		t.Fatalf("wanted std out buffer: hello kitteh out\n got std out buffer: %s", stdOutBufferString)
	}

	stdErrBufferString := stdErrBuff.String()
	if stdErrBufferString != "hello kitteh err\n" {
		t.Fatalf("wanted std err buffer: hello kitteh err\n got std err buffer: %s", stdErrBufferString)
	}
}
