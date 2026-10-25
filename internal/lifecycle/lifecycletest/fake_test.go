// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package lifecycletest_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
)

var (
	errBoom    = errors.New("boom")
	errDropped = errors.New("dropped")
)

func TestFakeDefaults(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	fake := lifecycletest.NewFake("a", rec)
	if err := fake.Start(t.Context()); err != nil {
		t.Errorf("Start() error = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := fake.Run(ctx); err != nil {
		t.Errorf("Run() after cancel error = %v", err)
	}
	fake.Fail(errBoom)
	fake.Fail(errDropped)
	if err := fake.Run(t.Context()); !errors.Is(err, errBoom) {
		t.Errorf("Run() after Fail error = %v, want %v", err, errBoom)
	}
	if err := fake.Stop(t.Context()); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	if got, want := rec.Calls(), []string{"a start", "a stop"}; !slices.Equal(got, want) {
		t.Errorf("Calls() = %q, want %q", got, want)
	}
}

func TestFakeFunctions(t *testing.T) {
	t.Parallel()
	rec := &lifecycletest.Recorder{}
	fake := lifecycletest.NewFake("a", rec)
	var called []string
	fake.StartFunc = func(context.Context) error {
		called = append(called, "start")
		return errBoom
	}
	fake.RunFunc = func(context.Context) error {
		called = append(called, "run")
		return errBoom
	}
	fake.StopFunc = func(context.Context) error {
		called = append(called, "stop")
		return errBoom
	}
	for name, fn := range map[string]func(context.Context) error{
		"Start": fake.Start, "Run": fake.Run, "Stop": fake.Stop,
	} {
		if err := fn(t.Context()); !errors.Is(err, errBoom) {
			t.Errorf("%s() error = %v, want %v", name, err, errBoom)
		}
	}
	slices.Sort(called)
	if want := []string{"run", "start", "stop"}; !slices.Equal(called, want) {
		t.Errorf("functions called = %q, want %q", called, want)
	}
}

func TestBlock(t *testing.T) {
	t.Parallel()
	entered, release := make(chan struct{}), make(chan struct{})
	block := lifecycletest.Block(entered, release)
	done := make(chan error)
	go func() { done <- block(t.Context()) }()
	<-entered
	close(release)
	if err := <-done; err != nil {
		t.Errorf("Block() after release error = %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := lifecycletest.Block(make(chan struct{}), make(chan struct{}))(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("Block() after cancel error = %v, want %v", err, context.Canceled)
	}
}
