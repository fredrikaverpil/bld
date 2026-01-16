package pocket

import (
	"context"
	"testing"
)

func TestRun_ExecutesCommand(t *testing.T) {
	// Run is hard to test without executing real commands,
	// so we test that it creates the right structure and skips in collect mode.

	cmd := Run("echo", "hello")

	// Should return nil funcs (leaf node)
	if funcs := cmd.funcs(); len(funcs) != 0 {
		t.Errorf("Run.funcs() = %v, want empty", funcs)
	}

	// Should skip in collect mode
	out := StdOutput()
	plan := newExecutionPlan()
	ec := &execContext{
		mode:  modeCollect,
		out:   out,
		cwd:   ".",
		dedup: newDedupState(),
		plan:  plan,
	}
	ctx := withExecContext(context.Background(), ec)

	if err := cmd.run(ctx); err != nil {
		t.Errorf("Run.run() in collect mode = %v, want nil", err)
	}
}

func TestRunWith_DynamicArgs(t *testing.T) {
	var capturedArgs []string

	// Create a RunWith that captures args for testing
	cmd := RunWith("test-cmd", func(ctx context.Context) []string {
		opts := Options[testOptions](ctx)
		args := []string{"base"}
		if opts.Extra != "" {
			args = append(args, opts.Extra)
		}
		capturedArgs = args
		return args
	})

	// Should return nil funcs (leaf node)
	if funcs := cmd.funcs(); len(funcs) != 0 {
		t.Errorf("RunWith.funcs() = %v, want empty", funcs)
	}

	// Should skip in collect mode (args function not called)
	out := StdOutput()
	plan := newExecutionPlan()
	ec := &execContext{
		mode:  modeCollect,
		out:   out,
		cwd:   ".",
		dedup: newDedupState(),
		plan:  plan,
	}
	ctx := withExecContext(context.Background(), ec)

	if err := cmd.run(ctx); err != nil {
		t.Errorf("RunWith.run() in collect mode = %v, want nil", err)
	}

	// Args function should NOT have been called in collect mode
	if capturedArgs != nil {
		t.Error("RunWith args function was called in collect mode, should be skipped")
	}
}

type testOptions struct {
	Extra string
}

func TestDo_ExecutesFunction(t *testing.T) {
	executed := false

	fn := Do(func(_ context.Context) error {
		executed = true
		return nil
	})

	// Should return nil funcs (leaf node)
	if funcs := fn.funcs(); len(funcs) != 0 {
		t.Errorf("Do.funcs() = %v, want empty", funcs)
	}

	// Should skip in collect mode
	out := StdOutput()
	plan := newExecutionPlan()
	ec := &execContext{
		mode:  modeCollect,
		out:   out,
		cwd:   ".",
		dedup: newDedupState(),
		plan:  plan,
	}
	ctx := withExecContext(context.Background(), ec)

	if err := fn.run(ctx); err != nil {
		t.Errorf("Do.run() in collect mode = %v, want nil", err)
	}

	if executed {
		t.Error("Do function was executed in collect mode, should be skipped")
	}

	// Should execute in execute mode
	ec = newExecContext(out, ".", false)
	ctx = withExecContext(context.Background(), ec)

	if err := fn.run(ctx); err != nil {
		t.Errorf("Do.run() in execute mode = %v, want nil", err)
	}

	if !executed {
		t.Error("Do function was not executed in execute mode")
	}
}

func TestDo_ComposesWithSerial(t *testing.T) {
	var order []string

	step1 := Do(func(_ context.Context) error {
		order = append(order, "step1")
		return nil
	})

	step2 := Do(func(_ context.Context) error {
		order = append(order, "step2")
		return nil
	})

	composed := Serial(step1, step2)

	out := StdOutput()
	ec := newExecContext(out, ".", false)
	ctx := withExecContext(context.Background(), ec)

	if err := composed.run(ctx); err != nil {
		t.Fatalf("Serial(Do, Do).run() = %v, want nil", err)
	}

	if len(order) != 2 {
		t.Errorf("expected 2 executions, got %d", len(order))
	}
	if order[0] != "step1" || order[1] != "step2" {
		t.Errorf("wrong execution order: %v", order)
	}
}

func TestRun_ComposesWithFuncDef(t *testing.T) {
	// Test that Run can be used as the body of a FuncDef
	// This will fail to execute (no such command) but tests composition

	taskFunc := Func("test-task", "run test command", Run("echo", "hello"))

	// Should have the FuncDef in funcs
	funcs := taskFunc.funcs()
	if len(funcs) != 1 {
		t.Errorf("expected 1 func, got %d", len(funcs))
	}
	if funcs[0].Name() != "test-task" {
		t.Errorf("expected name 'test-task', got %q", funcs[0].Name())
	}
}
