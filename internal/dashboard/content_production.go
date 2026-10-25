// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

//go:build dashboard_production

package dashboard

import "embed"

// content contains only the production output copied by the release build task.
//
//go:embed all:assets
var content embed.FS

// entry is the production entry page; a missing build fails compilation.
//
//go:embed assets/index.html
var entry []byte

const contentRoot = "assets"
