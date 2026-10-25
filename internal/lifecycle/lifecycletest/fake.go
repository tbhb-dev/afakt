// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package lifecycletest provides fake services for tests of code that uses the lifecycle
// package.
package lifecycletest

import (
	"context"
	"slices"
	"sync"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
)

var _ lifecycle.Service = (*Fake)(nil)

// Recorder collects the Start and Stop calls of fakes in the order they happen.
type Recorder struct {
	mu    sync.Mutex
	calls []string
}

// Calls returns the recorded calls in order. Each entry joins the fake's name and the method
// with a space, as in the entry db start for a Start call on the fake named db.
func (r *Recorder) Calls() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.calls)
}

func (r *Recorder) add(call string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, call)
}

// Fake is a service whose behavior a test sets. With no functions set, Start and Stop
// succeed at once, and Run waits until its context ends or Fail is called. Setting a
// function replaces the default behavior of its method. Assign them before the manager
// starts. Start and Stop record each call in the recorder before they do anything else.
type Fake struct {
	StartFunc func(ctx context.Context) error
	RunFunc   func(ctx context.Context) error
	StopFunc  func(ctx context.Context) error

	name     string
	recorder *Recorder
	fail     chan error
}

// NewFake returns a fake that records its calls in recorder under name.
func NewFake(name string, recorder *Recorder) *Fake {
	return &Fake{name: name, recorder: recorder, fail: make(chan error, 1)}
}

// Fail makes the default Run return err. Only the first call has an effect.
func (f *Fake) Fail(err error) {
	select {
	case f.fail <- err:
	default:
	}
}

func (f *Fake) Start(ctx context.Context) error {
	f.recorder.add(f.name + " start")
	if f.StartFunc != nil {
		return f.StartFunc(ctx)
	}
	return nil
}

func (f *Fake) Run(ctx context.Context) error {
	if f.RunFunc != nil {
		return f.RunFunc(ctx)
	}
	select {
	case <-ctx.Done():
		return nil
	case err := <-f.fail:
		return err
	}
}

func (f *Fake) Stop(ctx context.Context) error {
	f.recorder.add(f.name + " stop")
	if f.StopFunc != nil {
		return f.StopFunc(ctx)
	}
	return nil
}

// Block returns a function for StartFunc, RunFunc, or StopFunc. When called, that function
// closes entered and waits until release is closed or its context ends. A release makes it
// return nil, and an ended context makes it return the context's error.
func Block(entered chan<- struct{}, release <-chan struct{}) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		close(entered)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
