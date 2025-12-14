package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hrygo/dialecta/internal/config"
	"github.com/hrygo/dialecta/internal/debate"
)

// Runner orchestrates the CLI application execution
type Runner struct {
	ui       *UI
	input    *InputReader
	cfg      *config.Config
	stream   bool
	executor *debate.Executor
}

// NewRunner creates a new CLI runner
func NewRunner(cfg *config.Config, stream bool) *Runner {
	return &Runner{
		ui:       DefaultUI(),
		input:    DefaultInputReader(),
		cfg:      cfg,
		stream:   stream,
		executor: debate.NewExecutor(cfg),
	}
}

// RunWithOptions creates a runner with custom UI and input reader
func NewRunnerWithOptions(cfg *config.Config, stream bool, ui *UI, input *InputReader) *Runner {
	return &Runner{
		ui:       ui,
		input:    input,
		cfg:      cfg,
		stream:   stream,
		executor: debate.NewExecutor(cfg),
	}
}

// Run executes the debate with the given material
func (r *Runner) Run(ctx context.Context, material string) error {
	// Validate material
	if err := ValidateMaterial(material); err != nil {
		return err
	}

	// Print banner and config
	r.ui.PrintBanner()
	r.ui.PrintConfig(r.cfg)

	if r.stream {
		return r.runStreaming(ctx, material)
	}
	return r.runNonStreaming(ctx, material)
}

// runStreaming executes the debate in streaming mode using sequential display
func (r *Runner) runStreaming(ctx context.Context, material string) error {
	r.ui.PrintDebating()

	// Status tracking
	var (
		mu        sync.Mutex
		proStatus = "Thinking..."
		conStatus = "Thinking..."
		proDone   bool
		conDone   bool
		judgeDone bool
	)

	// Helper to update status line
	updateStatus := func() {
		// ANSI Clear Line + Carriage Return
		fmt.Printf("\r\033[K")
		if !proDone || !conDone {
			// Show status if not both finished
			fmt.Printf("⏳ Status: 🔵 Pro [%s] | 🔴 Con [%s]", proStatus, conStatus)
		} else if !judgeDone {
			// If both debated, show Judge status
			fmt.Printf("⏳ Status: ⚖️  Judge is deliberating...")
		}
	}

	// Spinner control
	var (
		stopSpinner func()
		spinnerOnce sync.Once
	)

	startJudgeSpinner := func() {
		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		stopSpinner = func() {
			spinnerOnce.Do(func() {
				close(stop)
				wg.Wait()
				fmt.Printf("\r\033[K") // Final clear
			})
		}

		// Immediate visual feedback to eliminate "cold wait"
		// Clear the "Pro [Ready] | Con [Ready]" status line immediately
		fmt.Printf("\r\033[K%s%s⏳ Status: ⚖️  Judge is deliberating...%s", ColorBrightYellow, ColorBold, ColorReset)

		go func() {
			defer wg.Done()
			chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			i := 0
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					// Redraw with spinner animation
					fmt.Printf("\r\033[K%s%s⏳ Status: ⚖️  Judge is deliberating... %s%s",
						ColorBrightYellow, ColorBold, chars[i%len(chars)], ColorReset)
					i++
				}
			}
		}()
	}

	// Initial status
	updateStatus()

	r.executor.SetStream(
		// Pro Callback
		func(chunk string, done bool) {
			mu.Lock()
			defer mu.Unlock()
			if done {
				proDone = true
				if conDone {
					// Both done, trigger judge spinner immediately
					startJudgeSpinner()
				}
				return
			}

			// We received the One-Liner
			// Clear status line
			fmt.Printf("\r\033[K")
			r.ui.PrintProHeader()
			fmt.Println(chunk)
			fmt.Println("") // Spacing

			proStatus = "Ready"
			updateStatus()
		},
		// Con Callback
		func(chunk string, done bool) {
			mu.Lock()
			defer mu.Unlock()
			if done {
				conDone = true
				if proDone {
					startJudgeSpinner()
				}
				return
			}

			fmt.Printf("\r\033[K")
			r.ui.PrintConHeader()
			fmt.Println(chunk)
			fmt.Println("")

			conStatus = "Ready"
			updateStatus()
		},
		// Judge Callback
		func(chunk string, done bool) {
			mu.Lock()
			defer mu.Unlock()

			// Stop spinner on first activity
			if stopSpinner != nil {
				stopSpinner()
			}

			if done {
				judgeDone = true
				return
			}

			fmt.Printf("\r\033[K")
			r.ui.PrintJudgeHeader()
			fmt.Println(chunk)
			fmt.Println("")
		},
	)

	result, err := r.executor.Execute(ctx, material)

	// Clear any remaining status line
	fmt.Printf("\r\033[K")

	if err != nil {
		r.ui.PrintError(err.Error())
		return err
	}

	// Final Summary
	fmt.Println()
	r.ui.PrintDivider()
	fmt.Printf("📄 Full Debate Report Saved: %s\n", result.ReportPath)
	r.ui.PrintDivider()

	r.ui.PrintComplete()

	return nil
}

// runNonStreaming executes the debate in non-streaming mode
func (r *Runner) runNonStreaming(ctx context.Context, material string) error {
	r.ui.PrintDebating()

	result, err := r.executor.Execute(ctx, material)
	if err != nil {
		return fmt.Errorf("执行失败: %w", err)
	}

	r.ui.PrintResult(result)
	r.ui.PrintComplete()

	return nil
}

// SetupContext creates a context that can be cancelled by interrupt signals
func SetupContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		ui := DefaultUI()
		ui.PrintWarning("中断信号接收，正在取消...")
		cancel()
	}()

	return ctx, cancel
}
