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
	// The keys are plain integers because exhaustive flags a map over a named integer
	// type unless every constant is listed.
	kinds := map[int]string{
		int(syscall.ENOSPC): "StorageFull",
		int(syscall.EPIPE):  "BrokenPipe",
	}
	code := int(errno)
	kind, ok := kinds[code]
	if !ok {
		kind = "Uncategorized"
	}
	message := errno.Error()
	message = strings.ToUpper(message[:1]) + message[1:]
	return fmt.Sprintf("Os { code: %d, kind: %s, message: %q }", code, kind, message)
}
