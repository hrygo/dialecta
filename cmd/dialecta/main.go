package main

import (
	"flag"
	"os"

	"github.com/hrygo/dialecta/internal/cli"
	"github.com/hrygo/dialecta/internal/config"
)

func main() {
	// Parse command-line flags
	opts := cli.ParseFlags()

	// Show help if needed
	if opts.NeedsHelp() {
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration and apply options
	cfg := config.New()
	opts.ApplyToConfig(cfg)

	// In interactive mode, let user select model combination
	if opts.Interactive {
		reader := cli.DefaultInputReader()
		combination, err := reader.SelectModelCombination()
		if err != nil {
			ui := cli.DefaultUI()
			ui.PrintError("选择模型组合失败: " + err.Error())
			os.Exit(1)
		}
		// Apply selected combination to config
		opts.JudgeProvider = combination.JudgeProvider
		opts.ProProvider = combination.ProProvider
		opts.ConProvider = combination.ConProvider
		opts.ApplyToConfig(cfg)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		ui := cli.DefaultUI()
		ui.PrintError("配置错误: " + err.Error())
		os.Exit(1)
	}

	// Read material
	reader := cli.DefaultInputReader()
	material, err := reader.ReadMaterial(opts.Source, opts.Interactive)
	if err != nil {
		ui := cli.DefaultUI()
		ui.PrintError("读取材料失败: " + err.Error())
		os.Exit(1)
	}

	// Setup context with signal handling
	ctx, cancel := cli.SetupContext()
	defer cancel()

	// Run the debate
	runner := cli.NewRunner(cfg, opts.Stream)
	if err := runner.Run(ctx, material); err != nil {
		ui := cli.DefaultUI()
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}
