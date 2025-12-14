package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/hrygo/dialecta/internal/config"
	"github.com/hrygo/dialecta/internal/llm"
)

// Options holds the parsed command-line options
type Options struct {
	ProProvider   string
	ProModel      string
	ConProvider   string
	ConModel      string
	JudgeProvider string
	JudgeModel    string
	Stream        bool
	Interactive   bool
	Source        string // file path, "-" for stdin, or empty for no source
}

// ParseFlags parses command-line flags and returns Options
func ParseFlags() *Options {
	opts := &Options{}

	flag.StringVar(&opts.ProProvider, "pro-provider", "deepseek", "Provider for affirmative (deepseek, gemini, dashscope)")
	flag.StringVar(&opts.ProModel, "pro-model", "", "Model for affirmative")
	flag.StringVar(&opts.ConProvider, "con-provider", "dashscope", "Provider for negative (deepseek, gemini, dashscope)")
	flag.StringVar(&opts.ConModel, "con-model", "", "Model for negative")
	flag.StringVar(&opts.JudgeProvider, "judge-provider", "gemini", "Provider for adjudicator (deepseek, gemini, dashscope)")
	flag.StringVar(&opts.JudgeModel, "judge-model", "", "Model for adjudicator")
	flag.BoolVar(&opts.Stream, "stream", true, "Enable streaming output")
	flag.BoolVar(&opts.Interactive, "interactive", false, "Interactive mode - enter material via stdin")
	flag.BoolVar(&opts.Interactive, "i", false, "Interactive mode (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
%s%s╭──────────────────────────────────────────────────────────────╮%s
%s%s│             DIALECTA - AI Debate Engine                      │%s
%s%s╰──────────────────────────────────────────────────────────────╯%s

%s%sUSAGE%s
  dialecta [options] <file>       %s▸ Analyze material from file%s
  dialecta [options] -            %s▸ Read from stdin (pipe)%s
  dialecta --interactive / -i     %s▸ Interactive input mode%s

%s%sAI PROVIDERS%s
  %s◈ deepseek%s   DeepSeek API       %s→ DEEPSEEK_API_KEY%s
  %s◈ gemini%s     Google Gemini      %s→ GEMINI_API_KEY / GOOGLE_API_KEY%s
  %s◈ dashscope%s  Alibaba Qwen       %s→ DASHSCOPE_API_KEY%s

%s%sEXAMPLES%s
  %s$%s dialecta proposal.md
  %s$%s cat plan.txt | dialecta -
  %s$%s echo "我们应该启动AI创业项目" | dialecta -
  %s$%s dialecta --judge-provider deepseek --judge-model deepseek-chat doc.md

%s%sOPTIONS%s
`, ColorBrightCyan, ColorBold, ColorReset,
			ColorBrightCyan, ColorBold, ColorReset,
			ColorBrightCyan, ColorBold, ColorReset,
			ColorBrightWhite, ColorBold, ColorReset,
			ColorDim, ColorReset,
			ColorDim, ColorReset,
			ColorDim, ColorReset,
			ColorBrightWhite, ColorBold, ColorReset,
			ColorBrightGreen, ColorReset, ColorDim, ColorReset,
			ColorBrightMagenta, ColorReset, ColorDim, ColorReset,
			ColorBrightYellow, ColorReset, ColorDim, ColorReset,
			ColorBrightWhite, ColorBold, ColorReset,
			ColorBrightCyan, ColorReset,
			ColorBrightCyan, ColorReset,
			ColorBrightCyan, ColorReset,
			ColorBrightCyan, ColorReset,
			ColorBrightWhite, ColorBold, ColorReset)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}

	flag.Parse()

	// Get source from remaining arguments
	if flag.NArg() > 0 {
		opts.Source = flag.Arg(0)
	}

	return opts
}

// ApplyToConfig applies the options to a config
// Temperature and MaxTokens are role-specific and remain unchanged when switching providers
func (opts *Options) ApplyToConfig(cfg *config.Config) {
	// Pro role - keep role-specific Temperature and MaxTokens
	if p, err := llm.ParseProvider(opts.ProProvider); err == nil {
		cfg.ProRole.Provider = p
		// If model not explicitly set, use provider's default
		if opts.ProModel == "" {
			cfg.ProRole.Model = config.GetDefaultModel(p)
		}
		// Temperature and MaxTokens remain as role defaults (0.8, 4096 for Pro)
	}
	if opts.ProModel != "" {
		cfg.ProRole.Model = opts.ProModel
	}

	// Con role - keep role-specific Temperature and MaxTokens
	if p, err := llm.ParseProvider(opts.ConProvider); err == nil {
		cfg.ConRole.Provider = p
		// If model not explicitly set, use provider's default
		if opts.ConModel == "" {
			cfg.ConRole.Model = config.GetDefaultModel(p)
		}
		// Temperature and MaxTokens remain as role defaults (0.8, 4096 for Con)
	}
	if opts.ConModel != "" {
		cfg.ConRole.Model = opts.ConModel
	}

	// Judge role - keep role-specific Temperature and MaxTokens
	if p, err := llm.ParseProvider(opts.JudgeProvider); err == nil {
		cfg.JudgeRole.Provider = p
		// If model not explicitly set, use provider's default
		if opts.JudgeModel == "" {
			cfg.JudgeRole.Model = config.GetDefaultModel(p)
		}
		// Temperature and MaxTokens remain as role defaults (0.1, 8192 for Judge)
	}
	if opts.JudgeModel != "" {
		cfg.JudgeRole.Model = opts.JudgeModel
	}
}

// NeedsHelp returns true if help should be shown (no source and not interactive)
func (opts *Options) NeedsHelp() bool {
	return opts.Source == "" && !opts.Interactive
}
