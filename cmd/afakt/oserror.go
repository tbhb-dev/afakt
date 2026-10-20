// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
)

// describeWriteError renders a write error as the text docs/cli.md documents after the
// Error label. An error from the operating system prints its code, an error kind name
// and the platform message with a capital first letter. Any other error prints as a
// custom error of kind Other.
func describeWriteError(err error) string {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return fmt.Sprintf("Custom { kind: Other, error: %q }", err.Error())
	}
	code := int(errno)
	kind, ok := errnoKinds()[code]
	if !ok {
		kind = "Uncategorized"
	}
	message := errno.Error()
	message = strings.ToUpper(message[:1]) + message[1:]
	return fmt.Sprintf("Os { code: %d, kind: %s, message: %q }", code, kind, message)
}

// errnoKinds maps an operating system error number to the name of the Rust io::ErrorKind
// that Rust's standard library gives it. EWOULDBLOCK has no entry because it equals EAGAIN
// on every platform Go supports, and a map literal rejects a duplicate key. The keys are
// plain integers because exhaustive flags a map over a named integer type unless every
// constant is listed.
func errnoKinds() map[int]string {
	return map[int]string{
		int(syscall.E2BIG):         "ArgumentListTooLong",
		int(syscall.EADDRINUSE):    "AddrInUse",
		int(syscall.EADDRNOTAVAIL): "AddrNotAvailable",
		int(syscall.EBUSY):         "ResourceBusy",
		int(syscall.ECONNABORTED):  "ConnectionAborted",
		int(syscall.ECONNREFUSED):  "ConnectionRefused",
		int(syscall.ECONNRESET):    "ConnectionReset",
		int(syscall.EDEADLK):       "Deadlock",
		int(syscall.EDQUOT):        "QuotaExceeded",
		int(syscall.EEXIST):        "AlreadyExists",
		int(syscall.EFBIG):         "FileTooLarge",
		int(syscall.EHOSTUNREACH):  "HostUnreachable",
		int(syscall.EINTR):         "Interrupted",
		int(syscall.EINVAL):        "InvalidInput",
		int(syscall.EISDIR):        "IsADirectory",
		int(syscall.ELOOP):         "FilesystemLoop",
		int(syscall.ENOENT):        "NotFound",
		int(syscall.ENOMEM):        "OutOfMemory",
		int(syscall.ENOSPC):        "StorageFull",
		int(syscall.ENOSYS):        "Unsupported",
		int(syscall.EMLINK):        "TooManyLinks",
		int(syscall.ENAMETOOLONG):  "InvalidFilename",
		int(syscall.ENETDOWN):      "NetworkDown",
		int(syscall.ENETUNREACH):   "NetworkUnreachable",
		int(syscall.ENOTCONN):      "NotConnected",
		int(syscall.ENOTDIR):       "NotADirectory",
		int(syscall.ENOTEMPTY):     "DirectoryNotEmpty",
		int(syscall.EPIPE):         "BrokenPipe",
		int(syscall.EROFS):         "ReadOnlyFilesystem",
		int(syscall.ESPIPE):        "NotSeekable",
		int(syscall.ESTALE):        "StaleNetworkFileHandle",
		int(syscall.ETIMEDOUT):     "TimedOut",
		int(syscall.ETXTBSY):       "ExecutableFileBusy",
		int(syscall.EXDEV):         "CrossesDevices",
		int(syscall.EINPROGRESS):   "InProgress",
		int(syscall.EOPNOTSUPP):    "Unsupported",
		int(syscall.EACCES):        "PermissionDenied",
		int(syscall.EPERM):         "PermissionDenied",
		int(syscall.EAGAIN):        "WouldBlock",
	}
}
