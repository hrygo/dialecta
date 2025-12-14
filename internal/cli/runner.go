package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

// runStreaming executes the debate in streaming mode using a dual-column layout
func (r *Runner) runStreaming(ctx context.Context, material string) error {
	r.ui.PrintDebating()

	// Phase 1: Dual-Column Parallel Stream (Pro/Con)
	dp := r.ui.StartDualStream()

	var judgeStarted bool

	r.executor.SetStream(
		func(chunk string) {
			dp.UpdatePro(chunk)
		},
		func(chunk string) {
			dp.UpdateCon(chunk)
		},
		func(chunk string) {
			// Phase 2: Judge Stream (Sequential after Pro/Con)
			if !judgeStarted {
				r.ui.Println("\n") // Spacer from columns
				r.ui.PrintJudgeHeader()
				judgeStarted = true
			}
			r.ui.Print(chunk)
		},
	)

	_, err := r.executor.Execute(ctx, material)
	if err != nil {
		return fmt.Errorf("执行失败: %w", err)
	}

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
