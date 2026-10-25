// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package lifecycle

// Status reports the state of one service to the external tests.
type Status struct {
	Name  string
	State State
}

// Statuses returns the state of each service in registration order. Only tests read the
// states, so this lives in a test file.
func (m *Manager) Statuses() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	statuses := make([]Status, 0, len(m.services))
	for _, s := range m.services {
		statuses = append(statuses, Status{Name: s.spec.Name, State: s.state})
	}
	return statuses
}
