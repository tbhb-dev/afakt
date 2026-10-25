// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package cmdutil holds what the command packages share. Besides the factory, its flag
// types, flag errors, and help layout keep the command line contract that the program had
// before it moved to Cobra, down to the wording of each error.
package cmdutil

import "io"

// Factory carries what a command takes from the process. Main fills it in. Commands read
// their streams and the version from it rather than from the os package.
type Factory struct {
	In      io.Reader
	Out     io.Writer
	ErrOut  io.Writer
	Version string
}
