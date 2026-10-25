// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package lifecycle_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
)

var (
	errBoom   = errors.New("boom")
	errIgnore = errors.New("ignored second failure")
)

// newManager returns a manager that discards its logs.
func newManager() *lifecycle.Manager {
	return lifecycle.New(slog.New(slog.DiscardHandler))
}

// mustRegister registers each registration in order and stops the test at the first error.
func mustRegister(t *testing.T, m *lifecycle.Manager, regs ...lifecycle.Registration) {
	t.Helper()
	for _, reg := range regs {
		if err := m.Register(reg); err != nil {
			t.Fatalf("Register(%q) error = %v", reg.Name, err)
		}
	}
}

// states maps the name of each registered service to its current state.
func states(m *lifecycle.Manager) map[string]lifecycle.State {
	got := map[string]lifecycle.State{}
	for _, s := range m.Statuses() {
		got[s.Name] = s.State
	}
	return got
}

// wantStates checks the state of each service the map names. Services missing from the
// map aren't checked.
func wantStates(t *testing.T, m *lifecycle.Manager, want map[string]lifecycle.State) {
	t.Helper()
	got := states(m)
	for name, state := range want {
		if got[name] != state {
			t.Errorf("state of %q = %s, want %s", name, got[name], state)
		}
	}
}

// wantCalls checks that the recorder logged exactly the expected calls, in that order.
func wantCalls(t *testing.T, rec *lifecycletest.Recorder, want ...string) {
	t.Helper()
	if got := rec.Calls(); !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

// checkedFake returns a fake that checks the ordering rules when they apply. Its Start
// reports an error unless every service in prerequisites is running. Its Stop reports an error
// while a service in dependents hasn't finished.
func checkedFake(t *testing.T, m *lifecycle.Manager, rec *lifecycletest.Recorder, name string,
	prerequisites, dependents []string,
) *lifecycletest.Fake {
	t.Helper()
	f := lifecycletest.NewFake(name, rec)
	f.StartFunc = func(context.Context) error {
		got := states(m)
		for _, dep := range prerequisites {
			if got[dep] != lifecycle.StateRunning {
				t.Errorf("%s started while its dependency %s was %s", name, dep, got[dep])
			}
		}
		return nil
	}
	f.StopFunc = func(context.Context) error {
		got := states(m)
		for _, dependent := range dependents {
			switch got[dependent] {
			case lifecycle.StateStarting, lifecycle.StateRunning, lifecycle.StateStopping:
				t.Errorf("%s stopped while its dependent %s was %s", name, dependent, got[dependent])
			case lifecycle.StateNew, lifecycle.StateTerminated, lifecycle.StateFailed:
				// A dependent that never started or has finished doesn't hold the service.
			}
		}
		return nil
	}
	return f
}

// stopCtx returns a context for Stop. Its deadline only matters if a service hangs.
func stopCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func TestStateString(t *testing.T) {
	t.Parallel()
	cases := map[lifecycle.State]string{
		lifecycle.StateNew:        "new",
		lifecycle.StateStarting:   "starting",
		lifecycle.StateRunning:    "running",
		lifecycle.StateStopping:   "stopping",
		lifecycle.StateTerminated: "terminated",
		lifecycle.StateFailed:     "failed",
		lifecycle.State(99):       "State(99)",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", int(state), got, want)
		}
	}
}

// TestDependencyOrder checks that services start in dependency order with ties broken by
// registration order, and stop in the reverse order.
func TestDependencyOrder(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		mustRegister(t, m,
			lifecycle.Registration{Name: "d", Service: lifecycletest.NewFake("d", rec)},
			lifecycle.Registration{
				Name: "c", Service: checkedFake(t, m, rec, "c", []string{"b", "a"}, nil),
				DependsOn: []string{"b", "a"},
			},
			lifecycle.Registration{
				Name: "b", Service: checkedFake(t, m, rec, "b", []string{"a"}, []string{"c"}),
				DependsOn: []string{"a"},
			},
			lifecycle.Registration{Name: "a", Service: checkedFake(t, m, rec, "a", nil, []string{"b", "c"})},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		wantStates(t, m, map[string]lifecycle.State{
			"a": lifecycle.StateRunning, "b": lifecycle.StateRunning,
			"c": lifecycle.StateRunning, "d": lifecycle.StateRunning,
		})
		if err := m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		wantCalls(t, rec, "d start", "a start", "b start", "c start", "c stop", "b stop", "a stop", "d stop")
		wantStates(t, m, map[string]lifecycle.State{
			"a": lifecycle.StateTerminated, "b": lifecycle.StateTerminated,
			"c": lifecycle.StateTerminated, "d": lifecycle.StateTerminated,
		})
	})
}

func TestRegisterRejects(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	fake := lifecycletest.NewFake("a", rec)
	cases := []struct {
		name string
		reg  lifecycle.Registration
		want error
	}{
		{name: "empty name", reg: lifecycle.Registration{Service: fake}, want: lifecycle.ErrInvalidRegistration},
		{name: "no service", reg: lifecycle.Registration{Name: "b"}, want: lifecycle.ErrInvalidRegistration},
		{name: "duplicate", reg: lifecycle.Registration{Name: "a", Service: fake}, want: lifecycle.ErrDuplicateService},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newManager()
			mustRegister(t, m, lifecycle.Registration{Name: "a", Service: fake})
			if err := m.Register(tc.reg); !errors.Is(err, tc.want) {
				t.Errorf("Register() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDependsOnIsCopied(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	m := newManager()
	deps := []string{"a"}
	mustRegister(t, m,
		lifecycle.Registration{Name: "b", Service: lifecycletest.NewFake("b", rec), DependsOn: deps},
		lifecycle.Registration{Name: "a", Service: lifecycletest.NewFake("a", rec)},
	)
	deps[0] = "missing"
	if err := m.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := m.Stop(stopCtx(t)); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	wantCalls(t, rec, "a start", "b start", "b stop", "a stop")
}

func TestRegisterAndStartAfterStart(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	m := newManager()
	mustRegister(t, m, lifecycle.Registration{Name: "a", Service: lifecycletest.NewFake("a", rec)})
	if err := m.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := m.Register(lifecycle.Registration{Name: "b", Service: lifecycletest.NewFake("b", rec)}); !errors.Is(
		err, lifecycle.ErrNotIdle,
	) {
		t.Errorf("Register() after Start error = %v, want %v", err, lifecycle.ErrNotIdle)
	}
	if err := m.Start(t.Context()); !errors.Is(err, lifecycle.ErrNotIdle) {
		t.Errorf("second Start() error = %v, want %v", err, lifecycle.ErrNotIdle)
	}
	if err := m.Stop(stopCtx(t)); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := m.Start(t.Context()); !errors.Is(err, lifecycle.ErrNotIdle) {
		t.Errorf("Start() after Stop error = %v, want %v", err, lifecycle.ErrNotIdle)
	}
}

func TestStartRejectsGraph(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		regs func(rec *lifecycletest.Recorder) []lifecycle.Registration
		want error
		text string
	}{
		{
			name: "unknown dependency",
			regs: func(rec *lifecycletest.Recorder) []lifecycle.Registration {
				return []lifecycle.Registration{
					{Name: "a", Service: lifecycletest.NewFake("a", rec), DependsOn: []string{"nope"}},
				}
			},
			want: lifecycle.ErrUnknownDependency,
			text: `service "a": unknown dependency "nope"`,
		},
		{
			name: "required on optional",
			regs: func(rec *lifecycletest.Recorder) []lifecycle.Registration {
				return []lifecycle.Registration{
					{Name: "o", Service: lifecycletest.NewFake("o", rec), Optional: true},
					{Name: "r", Service: lifecycletest.NewFake("r", rec), DependsOn: []string{"o"}},
				}
			},
			want: lifecycle.ErrOptionalDependency,
			text: `service "r": required service depends on an optional service "o"`,
		},
		{
			name: "cycle",
			regs: func(rec *lifecycletest.Recorder) []lifecycle.Registration {
				return []lifecycle.Registration{
					{Name: "free", Service: lifecycletest.NewFake("free", rec)},
					{Name: "a", Service: lifecycletest.NewFake("a", rec), DependsOn: []string{"b"}},
					{Name: "b", Service: lifecycletest.NewFake("b", rec), DependsOn: []string{"a"}},
				}
			},
			want: lifecycle.ErrDependencyCycle,
			text: "dependency cycle among a, b",
		},
		{
			name: "self",
			regs: func(rec *lifecycletest.Recorder) []lifecycle.Registration {
				return []lifecycle.Registration{
					{Name: "a", Service: lifecycletest.NewFake("a", rec), DependsOn: []string{"a"}},
				}
			},
			want: lifecycle.ErrDependencyCycle,
			text: "dependency cycle among a",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rec := &lifecycletest.Recorder{}
			m := newManager()
			mustRegister(t, m, tc.regs(rec)...)
			err := m.Start(t.Context())
			if !errors.Is(err, tc.want) || err.Error() != tc.text {
				t.Errorf("Start() error = %v, want %q wrapping %v", err, tc.text, tc.want)
			}
			if err = m.Stop(stopCtx(t)); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
			wantCalls(t, rec)
			for _, s := range m.Statuses() {
				if s.State != lifecycle.StateTerminated {
					t.Errorf("state of %q = %s, want terminated", s.Name, s.State)
				}
			}
		})
	}
}

func TestStatesWhileStartingAndStopping(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		// The slow service blocks first in Start and later in Stop. The test checks the states
		// at both points.
		slow := lifecycletest.NewFake("slow", rec)
		startEntered, startRelease := make(chan struct{}), make(chan struct{})
		slow.StartFunc = lifecycletest.Block(startEntered, startRelease)
		stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
		slow.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
		mustRegister(t, m,
			lifecycle.Registration{Name: "a", Service: lifecycletest.NewFake("a", rec)},
			lifecycle.Registration{Name: "slow", Service: slow, DependsOn: []string{"a"}},
		)
		started := make(chan error)
		go func() { started <- m.Start(t.Context()) }()
		<-startEntered
		wantStates(t, m, map[string]lifecycle.State{"a": lifecycle.StateRunning, "slow": lifecycle.StateStarting})
		close(startRelease)
		if err := <-started; err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		stopped := make(chan error)
		go func() { stopped <- m.Stop(stopCtx(t)) }()
		<-stopEntered
		wantStates(t, m, map[string]lifecycle.State{"a": lifecycle.StateRunning, "slow": lifecycle.StateStopping})
		close(stopRelease)
		if err := <-stopped; err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	})
}

// TestRequiredStartFailure checks that the shutdown after a failed start stops the started
// services in reverse dependency order, and that c, which comes after the failure, never
// starts.
func TestRequiredStartFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		bad := lifecycletest.NewFake("bad", rec)
		bad.StartFunc = func(context.Context) error { return errBoom }
		mustRegister(t, m,
			lifecycle.Registration{Name: "a", Service: checkedFake(t, m, rec, "a", nil, []string{"b", "c"})},
			lifecycle.Registration{
				Name: "b", Service: checkedFake(t, m, rec, "b", []string{"a"}, []string{"bad"}),
				DependsOn: []string{"a"},
			},
			lifecycle.Registration{Name: "bad", Service: bad, DependsOn: []string{"b"}},
			lifecycle.Registration{Name: "c", Service: lifecycletest.NewFake("c", rec), DependsOn: []string{"a"}},
		)
		err := m.Start(t.Context())
		if !errors.Is(err, errBoom) || err.Error() != `start "bad": boom` {
			t.Errorf("Start() error = %v, want start \"bad\": boom", err)
		}
		if err = m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		wantCalls(t, rec, "a start", "b start", "bad start", "b stop", "a stop")
		wantStates(t, m, map[string]lifecycle.State{
			"a": lifecycle.StateTerminated, "b": lifecycle.StateTerminated,
			"bad": lifecycle.StateFailed, "c": lifecycle.StateTerminated,
		})
	})
}

func TestOptionalStartFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		bad := lifecycletest.NewFake("bad", rec)
		bad.StartFunc = func(context.Context) error { return errBoom }
		mustRegister(t, m,
			lifecycle.Registration{Name: "bad", Service: bad, Optional: true},
			lifecycle.Registration{
				Name: "dependent", Service: lifecycletest.NewFake("dependent", rec),
				DependsOn: []string{"bad"}, Optional: true,
			},
			lifecycle.Registration{Name: "r", Service: lifecycletest.NewFake("r", rec)},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		wantStates(t, m, map[string]lifecycle.State{
			"bad": lifecycle.StateFailed, "dependent": lifecycle.StateFailed, "r": lifecycle.StateRunning,
		})
		if err := m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		wantCalls(t, rec, "bad start", "r start", "r stop")
	})
}

func TestRequiredRunFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		a := lifecycletest.NewFake("a", rec)
		mustRegister(t, m,
			lifecycle.Registration{Name: "a", Service: a},
			lifecycle.Registration{Name: "b", Service: lifecycletest.NewFake("b", rec), DependsOn: []string{"a"}},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		a.Fail(errBoom)
		a.Fail(errIgnore)
		err := m.Wait(t.Context())
		if !errors.Is(err, errBoom) || err.Error() != `run "a": boom` {
			t.Errorf("Wait() error = %v, want run \"a\": boom", err)
		}
		if err = m.Wait(t.Context()); !errors.Is(err, errBoom) {
			t.Errorf("second Wait() error = %v, want %v", err, errBoom)
		}
		// Stop still calls Stop on the failed service to release what its Start acquired.
		if err = m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		wantCalls(t, rec, "a start", "b start", "b stop", "a stop")
		wantStates(t, m, map[string]lifecycle.State{"a": lifecycle.StateFailed, "b": lifecycle.StateTerminated})
	})
}

func TestRunExitCountsAsFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		a := lifecycletest.NewFake("a", rec)
		a.RunFunc = func(context.Context) error { return nil }
		mustRegister(t, m, lifecycle.Registration{Name: "a", Service: a})
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if err := m.Wait(t.Context()); !errors.Is(err, lifecycle.ErrExited) {
			t.Errorf("Wait() error = %v, want %v", err, lifecycle.ErrExited)
		}
		if err := m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	})
}

func TestOptionalRunFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		o := lifecycletest.NewFake("o", rec)
		mustRegister(t, m,
			lifecycle.Registration{Name: "r", Service: lifecycletest.NewFake("r", rec)},
			lifecycle.Registration{Name: "o", Service: o, Optional: true},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		o.Fail(errBoom)
		synctest.Wait()
		wantStates(t, m, map[string]lifecycle.State{"o": lifecycle.StateFailed, "r": lifecycle.StateRunning})
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		if err := m.Wait(ctx); err != nil {
			t.Errorf("Wait() error = %v, want nil", err)
		}
		if err := m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	})
}

func TestStartStopsOnEarlierFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		// a fails only once b has entered Start, and b returns only once the manager has
		// recorded that failure, so c never starts.
		bEntered := make(chan struct{})
		a := lifecycletest.NewFake("a", rec)
		a.RunFunc = func(context.Context) error {
			<-bEntered
			return errBoom
		}
		b := lifecycletest.NewFake("b", rec)
		b.StartFunc = func(context.Context) error {
			close(bEntered)
			synctest.Wait()
			return nil
		}
		mustRegister(t, m,
			lifecycle.Registration{Name: "a", Service: a},
			lifecycle.Registration{Name: "b", Service: b, DependsOn: []string{"a"}},
			lifecycle.Registration{Name: "c", Service: lifecycletest.NewFake("c", rec), DependsOn: []string{"b"}},
		)
		if err := m.Start(t.Context()); !errors.Is(err, errBoom) {
			t.Errorf("Start() error = %v, want %v", err, errBoom)
		}
		if err := m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		wantCalls(t, rec, "a start", "b start", "b stop", "a stop")
	})
}

func TestStartInterruptedByContext(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	m := newManager()
	mustRegister(t, m, lifecycle.Registration{Name: "a", Service: lifecycletest.NewFake("a", rec)})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := m.Start(ctx)
	if !errors.Is(err, context.Canceled) || err.Error() != "start interrupted: context canceled" {
		t.Errorf("Start() error = %v, want start interrupted: context canceled", err)
	}
	if err = m.Stop(stopCtx(t)); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	wantCalls(t, rec)
}

func TestWaitAfterContextEnds(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		a := lifecycletest.NewFake("a", rec)
		mustRegister(t, m, lifecycle.Registration{Name: "a", Service: a})
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := m.Wait(ctx); err != nil {
			t.Errorf("Wait() before a failure error = %v, want nil", err)
		}
		a.Fail(errBoom)
		synctest.Wait()
		// With a failure on record and the context ended, Wait must pick the failure on
		// every call, not whichever case a select happens to choose.
		for range 20 {
			if err := m.Wait(ctx); !errors.Is(err, errBoom) {
				t.Fatalf("Wait() error = %v, want %v", err, errBoom)
			}
		}
		if err := m.Stop(stopCtx(t)); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	})
}

// TestRequiredFailureWhileStopping checks that Stop reports a required failure that Wait
// never returned: a fails while Stop waits on its dependent b, after Wait returned nil.
func TestRequiredFailureWhileStopping(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		a := lifecycletest.NewFake("a", rec)
		b := lifecycletest.NewFake("b", rec)
		stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
		b.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
		mustRegister(t, m,
			lifecycle.Registration{Name: "a", Service: a},
			lifecycle.Registration{Name: "b", Service: b, DependsOn: []string{"a"}},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := m.Wait(ctx); err != nil {
			t.Fatalf("Wait() error = %v, want nil", err)
		}
		stopped := make(chan error)
		go func() { stopped <- m.Stop(stopCtx(t)) }()
		<-stopEntered
		a.Fail(errBoom)
		synctest.Wait()
		close(stopRelease)
		if err := <-stopped; !errors.Is(err, errBoom) || err.Error() != `run "a": boom` {
			t.Errorf("Stop() error = %v, want run \"a\": boom", err)
		}
		wantCalls(t, rec, "a start", "b start", "b stop", "a stop")
		wantStates(t, m, map[string]lifecycle.State{"a": lifecycle.StateFailed, "b": lifecycle.StateTerminated})
	})
}

func TestPanicsBecomeErrors(t *testing.T) {
	t.Parallel()
	boom := func(context.Context) error { panic("boom") }
	cases := []struct {
		name  string
		setup func(f *lifecycletest.Fake)
		text  string
	}{
		{
			name:  "start",
			setup: func(f *lifecycletest.Fake) { f.StartFunc = boom },
			text:  `start "a": service panicked: boom`,
		},
		{
			name:  "run",
			setup: func(f *lifecycletest.Fake) { f.RunFunc = boom },
			text:  `run "a": service panicked: boom`,
		},
		{
			name:  "stop",
			setup: func(f *lifecycletest.Fake) { f.StopFunc = boom },
			text:  `stop "a": service panicked: boom`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				rec := &lifecycletest.Recorder{}
				m := newManager()
				a := lifecycletest.NewFake("a", rec)
				tc.setup(a)
				mustRegister(t, m, lifecycle.Registration{Name: "a", Service: a})
				err := m.Start(t.Context())
				if err == nil {
					synctest.Wait()
					// A canceled context makes Wait report a recorded failure or return nil.
					ctx, cancel := context.WithCancel(t.Context())
					cancel()
					err = m.Wait(ctx)
				}
				if err == nil {
					err = m.Stop(stopCtx(t))
				} else if stopErr := m.Stop(stopCtx(t)); stopErr != nil {
					t.Errorf("Stop() error = %v", stopErr)
				}
				if !errors.Is(err, lifecycle.ErrPanic) || err.Error() != tc.text {
					t.Errorf("error = %v, want %q", err, tc.text)
				}
				wantStates(t, m, map[string]lifecycle.State{"a": lifecycle.StateFailed})
			})
		})
	}
}

func TestStopErrors(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		stopFails := lifecycletest.NewFake("stop-fails", rec)
		stopFails.StopFunc = func(context.Context) error { return errBoom }
		runFails := lifecycletest.NewFake("run-fails", rec)
		runFails.RunFunc = func(ctx context.Context) error {
			<-ctx.Done()
			return errBoom
		}
		canceled := lifecycletest.NewFake("canceled", rec)
		canceled.RunFunc = func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}
		mustRegister(t, m,
			lifecycle.Registration{Name: "canceled", Service: canceled},
			lifecycle.Registration{Name: "run-fails", Service: runFails},
			lifecycle.Registration{Name: "stop-fails", Service: stopFails},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		// Each error marks its service failed, and the sequence goes on to the next service.
		err := m.Stop(stopCtx(t))
		want := "stop \"stop-fails\": boom\nrun \"run-fails\": boom"
		if !errors.Is(err, errBoom) || err.Error() != want {
			t.Errorf("Stop() error = %q, want %q", err, want)
		}
		wantCalls(t, rec, "canceled start", "run-fails start", "stop-fails start",
			"stop-fails stop", "run-fails stop", "canceled stop")
		wantStates(t, m, map[string]lifecycle.State{
			"canceled":   lifecycle.StateTerminated,
			"run-fails":  lifecycle.StateFailed,
			"stop-fails": lifecycle.StateFailed,
		})
		if err = m.Stop(stopCtx(t)); err != nil {
			t.Errorf("second Stop() error = %v, want nil", err)
		}
	})
}

func TestStopBeforeStart(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	m := newManager()
	mustRegister(t, m, lifecycle.Registration{Name: "a", Service: lifecycletest.NewFake("a", rec)})
	if err := m.Stop(stopCtx(t)); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	wantCalls(t, rec)
	wantStates(t, m, map[string]lifecycle.State{"a": lifecycle.StateTerminated})
}

func TestShutdownDeadline(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var logs bytes.Buffer
		m := lifecycle.New(slog.New(slog.NewTextHandler(&logs, nil)))
		rec := &lifecycletest.Recorder{}
		first := lifecycletest.NewFake("first", rec)
		firstEntered, firstRelease := make(chan struct{}), make(chan struct{})
		first.RunFunc = lifecycletest.Block(firstEntered, firstRelease)
		skipped := lifecycletest.NewFake("skipped", rec)
		skipped.StartFunc = func(context.Context) error { return errBoom }
		quick := lifecycletest.NewFake("quick", rec)
		quick.StopFunc = func(context.Context) error {
			time.Sleep(4 * time.Second)
			return nil
		}
		hung := lifecycletest.NewFake("hung", rec)
		hungEntered, hungRelease := make(chan struct{}), make(chan struct{})
		hung.StopFunc = lifecycletest.Block(hungEntered, hungRelease)
		mustRegister(t, m,
			lifecycle.Registration{Name: "first", Service: first},
			lifecycle.Registration{Name: "skipped", Service: skipped, Optional: true},
			lifecycle.Registration{Name: "hung", Service: hung},
			lifecycle.Registration{Name: "quick", Service: quick},
		)
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		// One deadline of five seconds covers the whole sequence. The quick service takes four of
		// them to stop, and the hung one uses up the rest, so the manager doesn't stop first.
		start := time.Now()
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		err := m.Stop(ctx)
		if elapsed := time.Since(start); elapsed != 5*time.Second {
			t.Errorf("Stop() took %s, want the 5s deadline", elapsed)
		}
		want := `shutdown deadline exceeded while stopping "hung"; not confirmed stopped: hung, first`
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) || err.Error() != want {
			t.Errorf("Stop() error = %q, want %q", err, want)
		}
		wantCalls(t, rec, "first start", "skipped start", "hung start", "quick start", "quick stop", "hung stop")
		wantStates(t, m, map[string]lifecycle.State{
			"first": lifecycle.StateRunning, "hung": lifecycle.StateStopping, "quick": lifecycle.StateTerminated,
		})
		if !strings.Contains(logs.String(), `msg="shutdown deadline exceeded" service=hung pending="[hung first]"`) {
			t.Errorf("logs lack the deadline report:\n%s", logs.String())
		}
		<-firstEntered
		close(firstRelease)
		<-hungEntered
		close(hungRelease)
	})
}

func TestShutdownDeadlineWhileRunDrains(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		rec := &lifecycletest.Recorder{}
		m := newManager()
		draining := lifecycletest.NewFake("draining", rec)
		release := make(chan struct{})
		draining.RunFunc = func(context.Context) error {
			<-release
			return nil
		}
		mustRegister(t, m, lifecycle.Registration{Name: "draining", Service: draining})
		if err := m.Start(t.Context()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		if err := m.Stop(ctx); !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Errorf("Stop() error = %v, want %v", err, lifecycle.ErrShutdownTimeout)
		}
		wantCalls(t, rec, "draining start", "draining stop")
		close(release)
	})
}
