// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

//go:build !dashboard_production

package dashboard

import "embed"

// content contains the committed page that keeps Go-only work independent of Node.
//
//go:embed standin
var content embed.FS

// entry is the entry page, embedded separately so it cannot fail to load at runtime.
//
//go:embed standin/index.html
var entry []byte

const contentRoot = "standin"
