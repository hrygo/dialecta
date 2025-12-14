package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hrygo/dialecta/internal/config"
	"github.com/hrygo/dialecta/internal/debate"
	"github.com/hrygo/dialecta/internal/llm"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func main() {
	// Flags
	proProvider := flag.String("pro-provider", "deepseek", "Provider for affirmative (deepseek, gemini, dashscope)")
	proModel := flag.String("pro-model", "", "Model for affirmative")
	conProvider := flag.String("con-provider", "dashscope", "Provider for negative (deepseek, gemini, dashscope)")
	conModel := flag.String("con-model", "", "Model for negative")
	judgeProvider := flag.String("judge-provider", "gemini", "Provider for adjudicator (deepseek, gemini, dashscope)")
	judgeModel := flag.String("judge-model", "", "Model for adjudicator")
	stream := flag.Bool("stream", true, "Enable streaming output")
	interactive := flag.Bool("interactive", false, "Interactive mode - enter material via stdin")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%sDialecta - Multi-Persona Debate System%s

Usage:
  dialecta [options] <file>       Analyze material from file
  dialecta [options] -            Read material from stdin (pipe)
  dialecta --interactive          Interactive mode

Providers:
  deepseek   - DeepSeek API (DEEPSEEK_API_KEY)
  gemini     - Google Gemini (GEMINI_API_KEY or GOOGLE_API_KEY)
  dashscope  - Alibaba DashScope/Qwen (DASHSCOPE_API_KEY)

Examples:
  dialecta proposal.md
  cat plan.txt | dialecta -
  echo "我们应该启动AI创业项目" | dialecta -
  dialecta --judge-provider deepseek --judge-model deepseek-chat proposal.md

Options:
`, colorBold, colorReset)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Load configuration
	cfg := config.New()

	// Override providers and models if specified
	if p, err := llm.ParseProvider(*proProvider); err == nil {
		cfg.ProRole.Provider = p
	}
	if *proModel != "" {
		cfg.ProRole.Model = *proModel
	}

	if p, err := llm.ParseProvider(*conProvider); err == nil {
		cfg.ConRole.Provider = p
	}
	if *conModel != "" {
		cfg.ConRole.Model = *conModel
	}

	if p, err := llm.ParseProvider(*judgeProvider); err == nil {
		cfg.JudgeRole.Provider = p
	}
	if *judgeModel != "" {
		cfg.JudgeRole.Model = *judgeModel
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s❌ 配置错误: %s%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Get material
	var material string
	var err error

	if *interactive {
		material, err = readInteractive()
	} else if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	} else if flag.Arg(0) == "-" {
		material, err = readStdin()
	} else {
		material, err = readFile(flag.Arg(0))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s❌ 读取材料失败: %s%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if strings.TrimSpace(material) == "" {
		fmt.Fprintf(os.Stderr, "%s❌ 材料内容为空%s\n", colorRed, colorReset)
		os.Exit(1)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintf(os.Stderr, "\n%s⚠️ 中断信号接收，正在取消...%s\n", colorYellow, colorReset)
		cancel()
	}()

	// Execute debate
	executor := debate.NewExecutor(cfg)

	fmt.Printf("\n%s╔══════════════════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
	fmt.Printf("%s║           🎭 Dialecta - 多角色辩论系统                        ║%s\n", colorCyan, colorReset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════╝%s\n\n", colorCyan, colorReset)

	// Print model info
	fmt.Printf("%s📋 配置信息%s\n", colorBold, colorReset)
	fmt.Printf("   正方: %s/%s\n", cfg.ProRole.Provider, cfg.ProRole.Model)
	fmt.Printf("   反方: %s/%s\n", cfg.ConRole.Provider, cfg.ConRole.Model)
	fmt.Printf("   裁决: %s/%s\n\n", cfg.JudgeRole.Provider, cfg.JudgeRole.Model)

	if *stream {
		// Streaming mode
		proStarted, conStarted := false, false

		executor.SetStream(
			func(chunk string) {
				if !proStarted {
					fmt.Printf("%s%s🟢 正方论述 (The Affirmative)%s\n", colorBold, colorGreen, colorReset)
					fmt.Println(strings.Repeat("─", 60))
					proStarted = true
				}
				fmt.Print(chunk)
			},
			func(chunk string) {
				if !conStarted {
					if proStarted {
						fmt.Println()
					}
					fmt.Printf("\n%s%s🔴 反方论述 (The Negative)%s\n", colorBold, colorRed, colorReset)
					fmt.Println(strings.Repeat("─", 60))
					conStarted = true
				}
				fmt.Print(chunk)
			},
			func(chunk string) {
				fmt.Print(chunk)
			},
		)

		fmt.Printf("%s⏳ 正反方并行辩论中...%s\n\n", colorYellow, colorReset)

		result, err := executor.Execute(ctx, material)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s❌ 执行失败: %s%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

		// Print verdict header (content already streamed)
		fmt.Printf("\n\n%s%s⚖️ 裁决方报告 (The Adjudicator)%s\n", colorBold, colorBlue, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(result.Verdict)

	} else {
		// Non-streaming mode
		fmt.Printf("%s⏳ 正反方并行辩论中...%s\n", colorYellow, colorReset)

		result, err := executor.Execute(ctx, material)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s❌ 执行失败: %s%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

		// Print results
		fmt.Printf("\n%s%s🟢 正方论述 (The Affirmative)%s\n", colorBold, colorGreen, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(result.ProArgument)

		fmt.Printf("\n%s%s🔴 反方论述 (The Negative)%s\n", colorBold, colorRed, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(result.ConArgument)

		fmt.Printf("\n%s%s⚖️ 裁决方报告 (The Adjudicator)%s\n", colorBold, colorBlue, colorReset)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(result.Verdict)
	}

	fmt.Printf("\n%s✅ 辩论完成%s\n", colorGreen, colorReset)
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readInteractive() (string, error) {
	fmt.Printf("%s📝 请输入待分析的材料（输入两个空行结束）:%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("─", 40))

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	emptyCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Trim trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n"), nil
}
