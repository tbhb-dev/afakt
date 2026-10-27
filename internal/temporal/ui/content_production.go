// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

//go:build dashboard_production

package ui

import (
	"io/fs"

	upstream "github.com/temporalio/ui-server/v2/ui"
)

// content embeds the pinned upstream distribution through its Go module, without runtime
// downloads or filesystem searches. Both directories exist in the module at compile time.
func content() fs.FS {
	return mustFiles(fs.Sub(mustFiles(upstream.Assets()), "local"))
}
