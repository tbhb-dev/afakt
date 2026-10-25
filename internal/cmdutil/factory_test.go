// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil_test

import (
	"fmt"
	"os"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

// A command writes to the streams of its factory, never to the process streams directly.
func ExampleFactory() {
	f := &cmdutil.Factory{Out: os.Stdout, Version: "1.2.3"}
	_, _ = fmt.Fprintf(f.Out, "afakt %s\n", f.Version)
	// Output: afakt 1.2.3
}
