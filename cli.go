package pocket

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"text/tabwriter"
)

// Main is the entry point for the CLI.
// It parses flags, handles -h/--help, and runs the specified task(s).
// If no task is specified, defaultTask is run.
func Main(tasks []*Task, defaultTask *Task) {
	os.Exit(run(tasks, defaultTask))
}

// run parses flags and runs tasks, returning the exit code.
func run(tasks []*Task, defaultTask *Task) int {
	verbose := flag.Bool("v", false, "verbose output")
	help := flag.Bool("h", false, "show help")
	flag.Usage = func() {
		printHelp(tasks, defaultTask)
	}
	flag.Parse()

	if *help {
		printHelp(tasks, defaultTask)
		return 0
	}

	// Build task map for lookup.
	taskMap := make(map[string]*Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.Name] = t
	}

	// Determine which tasks to run.
	args := flag.Args()
	var tasksToRun []*Task
	if len(args) == 0 {
		if defaultTask != nil {
			tasksToRun = []*Task{defaultTask}
		} else {
			fmt.Fprintln(os.Stderr, "no task specified and no default task")
			return 1
		}
	} else {
		for _, name := range args {
			t, ok := taskMap[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown task: %s\n", name)
				return 1
			}
			tasksToRun = append(tasksToRun, t)
		}
	}

	// Create context with cancellation on interrupt.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Set verbose mode in context.
	ctx = WithVerbose(ctx, *verbose)

	// Run the tasks.
	for _, t := range tasksToRun {
		if err := Run(ctx, t); err != nil {
			fmt.Fprintf(os.Stderr, "task %s failed: %v\n", t.Name, err)
			return 1
		}
	}
	return 0
}

// printHelp prints the help message with available tasks.
func printHelp(tasks []*Task, defaultTask *Task) {
	fmt.Println("Usage: pok [flags] [task...]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h    show help")
	fmt.Println("  -v    verbose output")
	fmt.Println()
	fmt.Println("Tasks:")

	// Sort tasks by name, excluding hidden ones.
	var visible []*Task
	for _, t := range tasks {
		if !t.Hidden {
			visible = append(visible, t)
		}
	}
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].Name < visible[j].Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, t := range visible {
		defaultMark := ""
		if defaultTask != nil && t.Name == defaultTask.Name {
			defaultMark = " (default)"
		}
		fmt.Fprintf(w, "  %s\t%s%s\n", t.Name, t.Usage, defaultMark)
	}
	w.Flush()
}
