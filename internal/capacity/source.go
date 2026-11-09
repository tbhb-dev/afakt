// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package capacity

import (
	"errors"
	"time"
)

var errUnsupportedSource = errors.New("unsupported capacity source")

// Parse dispatches by a closed source code. Layouts without fixtures stay unknown.
func Parse(source, pane string, now time.Time) ([]Reading, error) {
	switch source {
	case SourceClaudeUsage:
		return ParseClaudeUsage(pane, now), nil
	case SourceCodexStatus:
		return []Reading{unknown(VendorOpenAI, source, WindowOther, now)}, nil
	case SourceAGYUsage:
		return []Reading{unknown(VendorGoogle, source, WindowOther, now)}, nil
	default:
		return nil, errUnsupportedSource
	}
}
