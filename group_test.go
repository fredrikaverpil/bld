package pocket

import (
	"context"
	"testing"
)

func TestSerial_ExecutionMode(t *testing.T) {
	var executed []string

	fn1 := func(_ context.Context) error {
		executed = append(executed, "fn1")
		return nil
	}
	fn2 := func(_ context.Context) error {
		executed = append(executed, "fn2")
		return nil
	}

	// Create a FuncDef that uses Serial inside its body
	testFunc := Func("test", "test", func(_ context.Context) error {
		// This Serial call should find the implicit execution context
		Serial(fn1, fn2)
		return nil
	})

	// Create execution context and run
	out := StdOutput()
	ec := newExecContext(out, ".", false)
	ctx := withExecContext(context.Background(), ec)

	if err := testFunc.run(ctx); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(executed) != 2 {
		t.Errorf("expected 2 executions, got %d", len(executed))
	}
	if executed[0] != "fn1" || executed[1] != "fn2" {
		t.Errorf("wrong execution order: %v", executed)
	}
}

func TestParallel_ExecutionMode(t *testing.T) {
	executed := make(chan string, 2)

	fn1 := func(_ context.Context) error {
		executed <- "fn1"
		return nil
	}
	fn2 := func(_ context.Context) error {
		executed <- "fn2"
		return nil
	}

	// Create a FuncDef that uses Parallel inside its body
	testFunc := Func("test", "test", func(_ context.Context) error {
		// This Parallel call should find the implicit execution context
		Parallel(fn1, fn2)
		return nil
	})

	// Create execution context and run
	out := StdOutput()
	ec := newExecContext(out, ".", false)
	ctx := withExecContext(context.Background(), ec)

	if err := testFunc.run(ctx); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	close(executed)
	results := make([]string, 0, 2)
	for s := range executed {
		results = append(results, s)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 executions, got %d", len(results))
	}
}
