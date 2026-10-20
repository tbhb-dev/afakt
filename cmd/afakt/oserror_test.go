// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"errors"
	"io/fs"
	"syscall"
	"testing"
)

var errShort = errors.New("short write")

func TestDescribeWriteError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "storage full",
			err:  syscall.ENOSPC,
			want: `Os { code: 28, kind: StorageFull, message: "No space left on device" }`,
		},
		{
			name: "broken pipe",
			err:  syscall.EPIPE,
			want: `Os { code: 32, kind: BrokenPipe, message: "Broken pipe" }`,
		},
		{
			name: "uncategorized errno",
			err:  syscall.EIO,
			want: `Os { code: 5, kind: Uncategorized, message: "Input/output error" }`,
		},
		{
			name: "errno inside a path error",
			err:  &fs.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.ENOSPC},
			want: `Os { code: 28, kind: StorageFull, message: "No space left on device" }`,
		},
		{
			name: "error without an errno",
			err:  errShort,
			want: `Custom { kind: Other, error: "short write" }`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := describeWriteError(tc.err); got != tc.want {
				t.Errorf("describeWriteError() = %s, want %s", got, tc.want)
			}
		})
	}
}
