// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
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

type errnoKind struct {
	errno syscall.Errno
	kind  string
}

func wantKinds() []errnoKind {
	return []errnoKind{
		{syscall.E2BIG, "ArgumentListTooLong"},
		{syscall.EADDRINUSE, "AddrInUse"},
		{syscall.EADDRNOTAVAIL, "AddrNotAvailable"},
		{syscall.EBUSY, "ResourceBusy"},
		{syscall.ECONNABORTED, "ConnectionAborted"},
		{syscall.ECONNREFUSED, "ConnectionRefused"},
		{syscall.ECONNRESET, "ConnectionReset"},
		{syscall.EDEADLK, "Deadlock"},
		{syscall.EDQUOT, "QuotaExceeded"},
		{syscall.EEXIST, "AlreadyExists"},
		{syscall.EFBIG, "FileTooLarge"},
		{syscall.EHOSTUNREACH, "HostUnreachable"},
		{syscall.EINTR, "Interrupted"},
		{syscall.EINVAL, "InvalidInput"},
		{syscall.EISDIR, "IsADirectory"},
		{syscall.ELOOP, "FilesystemLoop"},
		{syscall.ENOENT, "NotFound"},
		{syscall.ENOMEM, "OutOfMemory"},
		{syscall.ENOSPC, "StorageFull"},
		{syscall.ENOSYS, "Unsupported"},
		{syscall.EMLINK, "TooManyLinks"},
		{syscall.ENAMETOOLONG, "InvalidFilename"},
		{syscall.ENETDOWN, "NetworkDown"},
		{syscall.ENETUNREACH, "NetworkUnreachable"},
		{syscall.ENOTCONN, "NotConnected"},
		{syscall.ENOTDIR, "NotADirectory"},
		{syscall.ENOTEMPTY, "DirectoryNotEmpty"},
		{syscall.EPIPE, "BrokenPipe"},
		{syscall.EROFS, "ReadOnlyFilesystem"},
		{syscall.ESPIPE, "NotSeekable"},
		{syscall.ESTALE, "StaleNetworkFileHandle"},
		{syscall.ETIMEDOUT, "TimedOut"},
		{syscall.ETXTBSY, "ExecutableFileBusy"},
		{syscall.EXDEV, "CrossesDevices"},
		{syscall.EINPROGRESS, "InProgress"},
		{syscall.EOPNOTSUPP, "Unsupported"},
		{syscall.EACCES, "PermissionDenied"},
		{syscall.EPERM, "PermissionDenied"},
		{syscall.EAGAIN, "WouldBlock"},
		{syscall.EWOULDBLOCK, "WouldBlock"},
		{syscall.EIO, "Uncategorized"},
		{syscall.EBADF, "Uncategorized"},
	}
}

func TestDescribeWriteErrorKinds(t *testing.T) {
	t.Parallel()
	for _, tc := range wantKinds() {
		errno, kind := tc.errno, tc.kind
		t.Run(kind+"/"+errno.Error(), func(t *testing.T) {
			t.Parallel()
			message := errno.Error()
			message = strings.ToUpper(message[:1]) + message[1:]
			want := fmt.Sprintf("Os { code: %d, kind: %s, message: %q }", int(errno), kind, message)
			if got := describeWriteError(errno); got != want {
				t.Errorf("describeWriteError(%d) = %s, want %s", int(errno), got, want)
			}
		})
	}
}

func TestErrnoKindsAreAllTested(t *testing.T) {
	t.Parallel()
	want := make(map[int]string)
	for _, tc := range wantKinds() {
		want[int(tc.errno)] = tc.kind
	}
	mapped := errnoKinds()
	for code, kind := range mapped {
		if got, ok := want[code]; !ok || got != kind {
			t.Errorf("errnoKinds()[%d] = %q, test table has %q (present: %t)", code, kind, got, ok)
		}
	}
	for code, kind := range want {
		if _, ok := mapped[code]; !ok && kind != "Uncategorized" {
			t.Errorf("errno %d (%s) is in the test table but not mapped", code, kind)
		}
	}
}
