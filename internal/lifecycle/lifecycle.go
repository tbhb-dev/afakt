// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package lifecycle starts a set of services in dependency order and later stops them in
// reverse order under one shutdown deadline. While they run, a failure ends the wait for a
// signal. The package imports only the standard library.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// Service is one component that the manager runs. Start acquires resources and returns
// once the service can do its work. If Start fails, it releases what it acquired. Run does
// the work until the manager cancels its context, and returning earlier counts as a
// failure. Stop releases what Start acquired, within the deadline of its context. The
// manager calls Stop once for each service whose Start succeeded, even when Run failed.
type Service interface {
	Start(ctx context.Context) error
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Registration describes one service for Register. DependsOn names the services that must
// be running before this one starts. A required service can depend only on other required
// services. When a required service fails, the manager reports a failure. When an optional
// one fails, it logs the failure and goes on.
type Registration struct {
	Name      string
	Service   Service
	DependsOn []string
	Optional  bool
}

// State is where a service is in its lifecycle.
type State int

// The states of a service. A service moves from StateNew to StateStarting to StateRunning,
// then to StateStopping and StateTerminated. From starting, running, or stopping it can
// move to StateFailed instead. A service that never started goes from StateNew straight to
// StateTerminated when the manager stops, or to StateFailed when a dependency isn't running.
const (
	StateNew State = iota
	StateStarting
	StateRunning
	StateStopping
	StateTerminated
	StateFailed
)

func (s State) String() string {
	switch s {
	case StateNew:
		return "new"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateTerminated:
		return "terminated"
	case StateFailed:
		return "failed"
	default:
		return "State(" + strconv.Itoa(int(s)) + ")"
	}
}

var (
	// ErrInvalidRegistration reports a registration that lacks a name or a service.
	ErrInvalidRegistration = errors.New("invalid service registration")
	// ErrDuplicateService reports a second registration with a name already registered.
	ErrDuplicateService = errors.New("duplicate service")
	// ErrNotIdle reports a call to Register or Start after Start.
	ErrNotIdle = errors.New("manager has already started")
	// ErrUnknownDependency reports a dependency on a service that isn't registered.
	ErrUnknownDependency = errors.New("unknown dependency")
	// ErrOptionalDependency reports a required service that depends on an optional one.
	ErrOptionalDependency = errors.New("required service depends on an optional service")
	// ErrDependencyCycle reports services that depend on each other.
	ErrDependencyCycle = errors.New("dependency cycle")
	// ErrDependencyNotRunning reports a service skipped because a dependency isn't running.
	ErrDependencyNotRunning = errors.New("dependency is not running")
	// ErrExited reports a Run that returned before the manager stopped the service.
	ErrExited = errors.New("service exited before it was stopped")
	// ErrPanic reports a panic in Start, Run, or Stop.
	ErrPanic = errors.New("service panicked")
	// ErrShutdownTimeout reports a stop sequence that ran past its deadline.
	ErrShutdownTimeout = errors.New("shutdown deadline exceeded")
)

// phase is where the manager is in its own lifecycle. Register and Start work only in
// the idle phase, and once Stop moves the manager to the stopping phase it never leaves.
type phase int

const (
	phaseIdle phase = iota
	phaseStarted
	phaseStopping
)

// service is the manager's record of one registered service. started reports that Start
// succeeded, so Stop is owed. stopping tells run that a return from Run is the requested
// stop rather than a failure. run stores Run's result before it closes the done channel.
type service struct {
	reg       Registration
	state     State
	started   bool
	stopping  bool
	cancelRun context.CancelFunc
	runDone   chan struct{}
	runErr    error
}

// Manager runs registered services. Call Register for each service, then Start, then Wait,
// then Stop.
type Manager struct {
	logger *slog.Logger

	mu       sync.Mutex
	phase    phase
	services []*service
	order    []*service
	failed   chan struct{}
	failure  error
}

// New returns a manager that logs through logger.
func New(logger *slog.Logger) *Manager {
	return &Manager{logger: logger, failed: make(chan struct{})}
}

// Register adds a service. The manager starts services in dependency order and breaks ties
// in registration order.
func (m *Manager) Register(reg Registration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch {
	case m.phase != phaseIdle:
		return fmt.Errorf("register %q: %w", reg.Name, ErrNotIdle)
	case reg.Name == "" || reg.Service == nil:
		return fmt.Errorf("register %q: %w", reg.Name, ErrInvalidRegistration)
	case m.find(reg.Name) != nil:
		return fmt.Errorf("register %q: %w", reg.Name, ErrDuplicateService)
	}
	reg.DependsOn = slices.Clone(reg.DependsOn)
	m.services = append(m.services, &service{reg: reg, runDone: make(chan struct{})})
	return nil
}

// find returns the registered service named name, or nil. The caller holds m.mu.
func (m *Manager) find(name string) *service {
	for _, s := range m.services {
		if s.reg.Name == name {
			return s
		}
	}
	return nil
}

// resolve checks the dependency graph and returns the start order. Each pass takes the
// first service in registration order whose dependencies are all placed. The caller holds
// m.mu.
func (m *Manager) resolve() ([]*service, error) {
	for _, s := range m.services {
		for _, name := range s.reg.DependsOn {
			dep := m.find(name)
			switch {
			case dep == nil:
				return nil, fmt.Errorf("service %q: %w %q", s.reg.Name, ErrUnknownDependency, name)
			case !s.reg.Optional && dep.reg.Optional:
				return nil, fmt.Errorf("service %q: %w %q", s.reg.Name, ErrOptionalDependency, name)
			}
		}
	}
	placed := make(map[string]bool, len(m.services))
	order := make([]*service, 0, len(m.services))
	for len(order) < len(m.services) {
		next := m.nextReady(placed)
		if next == nil {
			return nil, fmt.Errorf("%w among %s", ErrDependencyCycle, m.unplacedNames(placed))
		}
		placed[next.reg.Name] = true
		order = append(order, next)
	}
	return order, nil
}

// nextReady returns the first unplaced service in registration order whose dependencies
// are all placed, or nil when no such service is left. The caller holds m.mu.
func (m *Manager) nextReady(placed map[string]bool) *service {
	for _, s := range m.services {
		if placed[s.reg.Name] {
			continue
		}
		if !slices.ContainsFunc(s.reg.DependsOn, func(name string) bool { return !placed[name] }) {
			return s
		}
	}
	return nil
}

// unplacedNames lists the services that resolve couldn't place, for the cycle error. The
// caller holds m.mu.
func (m *Manager) unplacedNames(placed map[string]bool) string {
	var names []string
	for _, s := range m.services {
		if !placed[s.reg.Name] {
			names = append(names, s.reg.Name)
		}
	}
	return strings.Join(names, ", ")
}

// Start starts every service in dependency order. The result is the first failure of a
// required service, or the context's error if the context ends first. After Start returns,
// call Stop whatever the result, to release what the started services acquired.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.phase != phaseIdle {
		m.mu.Unlock()
		return fmt.Errorf("start: %w", ErrNotIdle)
	}
	order, err := m.resolve()
	if err != nil {
		m.mu.Unlock()
		return err
	}
	m.phase = phaseStarted
	m.order = order
	m.mu.Unlock()

	for _, s := range order {
		if err = m.interrupted(ctx); err != nil {
			return err
		}
		if err = m.startService(ctx, s); err != nil && !s.reg.Optional {
			return err
		}
	}
	return nil
}

// interrupted reports why startup should stop before the next service: an earlier required
// service failed, or the context ended.
func (m *Manager) interrupted(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failure != nil {
		return m.failure
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("start interrupted: %w", err)
	}
	return nil
}

// startService calls Start once every dependency is running, then runs Run in its own
// goroutine. A dependency that isn't running fails the service without a call to Start.
// It takes m.mu itself and releases it around the call to Start.
func (m *Manager) startService(ctx context.Context, s *service) error {
	name := s.reg.Name
	m.mu.Lock()
	for _, dep := range s.reg.DependsOn {
		if m.find(dep).state != StateRunning {
			err := fmt.Errorf("start %q: %w: %q", name, ErrDependencyNotRunning, dep)
			m.setFailed(ctx, s, err)
			m.mu.Unlock()
			return err
		}
	}
	m.setState(ctx, s, StateStarting)
	m.mu.Unlock()

	err := protect(func() error { return s.reg.Service.Start(ctx) })

	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		err = fmt.Errorf("start %q: %w", name, err)
		m.setFailed(ctx, s, err)
		return err
	}
	s.started = true
	m.setState(ctx, s, StateRunning)
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.cancelRun = cancel
	go m.run(runCtx, s)
	return nil
}

// run calls Run and records its result. A return before the manager stops the service is a
// failure, and a failure of a required service ends Wait.
func (m *Manager) run(ctx context.Context, s *service) {
	defer close(s.runDone)
	err := protect(func() error { return s.reg.Service.Run(ctx) })
	m.mu.Lock()
	defer m.mu.Unlock()
	s.runErr = err
	if s.stopping {
		return
	}
	if err == nil {
		err = ErrExited
	}
	err = fmt.Errorf("run %q: %w", s.reg.Name, err)
	m.setFailed(ctx, s, err)
	if !s.reg.Optional && m.failure == nil {
		m.failure = err
		close(m.failed)
	}
}

// Wait blocks until the context ends or a required service fails. It returns nil in the
// first case and the service's error in the second.
func (m *Manager) Wait(ctx context.Context) error {
	select {
	case <-m.failed:
		return m.failureErr()
	default:
	}
	select {
	case <-m.failed:
		return m.failureErr()
	case <-ctx.Done():
		return nil
	}
}

// failureErr returns the first failure of a required service, which is set before m.failed
// closes.
func (m *Manager) failureErr() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failure
}

// Stop stops the started services in reverse start order. The deadline of the context bounds
// the whole sequence. Once it passes, Stop gives up waiting and returns ErrShutdownTimeout
// with the services it couldn't confirm stopped. A second call returns nil at once.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if m.phase == phaseStopping {
		m.mu.Unlock()
		return nil
	}
	m.phase = phaseStopping
	// Before a successful resolve there is no start order, and no service has started.
	order := slices.Clone(m.order)
	if order == nil {
		order = slices.Clone(m.services)
	}
	slices.Reverse(order)
	m.mu.Unlock()

	var errs []error
	for i, s := range order {
		if !s.started {
			m.mu.Lock()
			if s.state == StateNew {
				m.setState(ctx, s, StateTerminated)
			}
			m.mu.Unlock()
			continue
		}
		timedOut, err := m.stopService(ctx, s)
		if timedOut {
			errs = append(errs, m.timeout(ctx, order[i:]))
			break
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// stopService cancels Run and calls Stop, then waits until both return. The first result
// reports whether the deadline passed first.
func (m *Manager) stopService(ctx context.Context, s *service) (bool, error) {
	name := s.reg.Name
	m.mu.Lock()
	s.stopping = true
	alreadyFailed := s.state == StateFailed
	if !alreadyFailed {
		m.setState(ctx, s, StateStopping)
	}
	m.mu.Unlock()
	s.cancelRun()

	stopped := make(chan error, 1)
	go func() { stopped <- protect(func() error { return s.reg.Service.Stop(ctx) }) }()
	var stopErr error
	select {
	case stopErr = <-stopped:
		select {
		case <-s.runDone:
		case <-ctx.Done():
		}
	case <-ctx.Done():
	}
	// A service that finished only as the deadline passed counts as timed out too, so the
	// report doesn't depend on which channel the select saw first.
	if ctx.Err() != nil {
		return true, nil
	}

	var errs []error
	if stopErr != nil {
		errs = append(errs, fmt.Errorf("stop %q: %w", name, stopErr))
	}
	if s.runErr != nil && !alreadyFailed && !errors.Is(s.runErr, context.Canceled) {
		errs = append(errs, fmt.Errorf("run %q: %w", name, s.runErr))
	}
	err := errors.Join(errs...)
	m.mu.Lock()
	defer m.mu.Unlock()
	switch {
	case err != nil:
		m.setFailed(ctx, s, err)
	case !alreadyFailed:
		m.setState(ctx, s, StateTerminated)
	}
	return false, err
}

// timeout logs and returns the error for a stop sequence that ran out of time. The first of
// pending is the service whose stop didn't finish.
func (m *Manager) timeout(ctx context.Context, pending []*service) error {
	names := make([]string, 0, len(pending))
	for _, s := range pending {
		if s.started {
			names = append(names, s.reg.Name)
		}
	}
	err := fmt.Errorf("%w while stopping %q; not confirmed stopped: %s",
		ErrShutdownTimeout, pending[0].reg.Name, strings.Join(names, ", "))
	m.logger.ErrorContext(ctx, "shutdown deadline exceeded", "service", pending[0].reg.Name, "pending", names)
	return err
}

// setState records a transition and logs it. The caller holds m.mu.
func (m *Manager) setState(ctx context.Context, s *service, state State) {
	s.state = state
	m.logger.InfoContext(ctx, "service "+state.String(), "service", s.reg.Name)
}

// setFailed records a failure and logs it. The caller holds m.mu.
func (m *Manager) setFailed(ctx context.Context, s *service, err error) {
	s.state = StateFailed
	level := slog.LevelError
	if s.reg.Optional {
		level = slog.LevelWarn
	}
	m.logger.Log(ctx, level, "service failed", "service", s.reg.Name, "required", !s.reg.Optional, "error", err)
}

// protect runs call and turns a panic into an error that wraps ErrPanic.
func protect(call func() error) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w: %v", ErrPanic, r)
			}
		}()
		err = call()
	}()
	return err
}
