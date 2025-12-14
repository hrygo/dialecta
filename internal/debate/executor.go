package debate

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hrygo/dialecta/internal/config"
	"github.com/hrygo/dialecta/internal/llm"
	"github.com/hrygo/dialecta/internal/prompt"
)

// Result holds the complete debate result
type Result struct {
	Material        string // 原始材料
	ProOneLiner     string // 正方一句话观点
	ProFullBody     string // 正方完整论述
	ConOneLiner     string // 反方一句话观点
	ConFullBody     string // 反方完整论述
	VerdictOneLiner string // 裁决一句话
	VerdictFullBody string // 裁决完整报告
	ReportPath      string // 报告文件路径
}

// Executor orchestrates the debate process
type Executor struct {
	cfg          *config.Config
	stream       bool
	onPro        func(string, bool) // (content, done)
	onCon        func(string, bool)
	onJudge      func(string, bool)
	onJudgeStart func() // Called right before judge phase begins
}

// NewExecutor creates a new debate executor
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		cfg:    cfg,
		stream: false,
	}
}

// SetStream enables streaming mode with callbacks
func (e *Executor) SetStream(onPro, onCon, onJudge func(string, bool)) {
	e.stream = true
	e.onPro = onPro
	e.onCon = onCon
	e.onJudge = onJudge
}

// SetJudgeStartCallback sets callback for when judge phase begins
func (e *Executor) SetJudgeStartCallback(onJudgeStart func()) {
	e.onJudgeStart = onJudgeStart
}

// Execute runs the full debate workflow
func (e *Executor) Execute(ctx context.Context, material string) (*Result, error) {
	result := &Result{Material: material}

	// Phase 1: 并行执行正反方辩论
	var wg sync.WaitGroup
	var proErr, conErr error

	proParser := NewStreamParser("## 📝 Full Argument")
	conParser := NewStreamParser("## 📝 Full Argument")

	wg.Add(2)

	// 正方
	go func() {
		defer wg.Done()
		client, err := llm.NewClient(e.cfg.ProRole.ToLLMConfig())
		if err != nil {
			proErr = fmt.Errorf("create pro client: %w", err)
			return
		}

		messages := prompt.BuildAffirmativeMessages(material)
		if e.stream && e.onPro != nil {
			// Stream Callback
			_, err = client.ChatStream(ctx, messages, func(chunk string) {
				// Feed parser
				oneLiner, found := proParser.Feed(chunk)
				if found {
					// Notify CLI with the One-Liner ONLY once
					e.onPro(oneLiner, false)
				}
			})
			proParser.Finalize()
			result.ProOneLiner = proParser.oneLiner
			result.ProFullBody = proParser.fullBody

			// Determine what to save if parsing failed (fallback)
			if result.ProFullBody == "" {
				result.ProFullBody = proParser.buffer.String()
			}

			e.onPro("", true) // Signal done
		} else {
			// Non-streaming logic (Simplified for now, assumes streaming is primary)
			full, err := client.Chat(ctx, messages)
			if err == nil {
				// We still parse for result structure
				proParser.Feed(full)
				proParser.Finalize()
				result.ProOneLiner = proParser.oneLiner
				result.ProFullBody = proParser.fullBody
			}
			proErr = err
		}
		if err != nil {
			proErr = fmt.Errorf("affirmative: %w", err)
		}
	}()

	// 反方
	go func() {
		defer wg.Done()
		client, err := llm.NewClient(e.cfg.ConRole.ToLLMConfig())
		if err != nil {
			conErr = fmt.Errorf("create con client: %w", err)
			return
		}

		messages := prompt.BuildNegativeMessages(material)
		if e.stream && e.onCon != nil {
			_, err = client.ChatStream(ctx, messages, func(chunk string) {
				oneLiner, found := conParser.Feed(chunk)
				if found {
					e.onCon(oneLiner, false)
				}
			})
			conParser.Finalize()
			result.ConOneLiner = conParser.oneLiner
			result.ConFullBody = conParser.fullBody
			if result.ConFullBody == "" {
				result.ConFullBody = conParser.buffer.String()
			}

			e.onCon("", true)
		} else {
			full, err := client.Chat(ctx, messages)
			if err == nil {
				conParser.Feed(full)
				conParser.Finalize()
				result.ConOneLiner = conParser.oneLiner
				result.ConFullBody = conParser.fullBody
			}
			conErr = err
		}
		if err != nil {
			conErr = fmt.Errorf("negative: %w", err)
		}
	}()

	wg.Wait()

	if proErr != nil {
		return nil, proErr
	}
	if conErr != nil {
		return nil, conErr
	}

	// Notify that judge phase is starting (before any preparation work)
	if e.onJudgeStart != nil {
		e.onJudgeStart()
	}

	// Phase 2: 裁决
	judgeClient, err := llm.NewClient(e.cfg.JudgeRole.ToLLMConfig())
	if err != nil {
		return nil, fmt.Errorf("create judge client: %w", err)
	}

	// Use Full Bodies for Judge context
	messages := prompt.BuildAdjudicatorMessages(material, result.ProFullBody, result.ConFullBody)
	judgeParser := NewStreamParser("## 📝 Full Verdict")

	if e.stream && e.onJudge != nil {
		_, _ = judgeClient.ChatStream(ctx, messages, func(chunk string) {
			oneLiner, found := judgeParser.Feed(chunk)
			if found {
				e.onJudge(oneLiner, false)
			}
		})
		judgeParser.Finalize()
		result.VerdictOneLiner = judgeParser.oneLiner
		result.VerdictFullBody = judgeParser.fullBody
		if result.VerdictFullBody == "" {
			result.VerdictFullBody = judgeParser.buffer.String()
		}

		e.onJudge("", true)
	} else {
		full, err := judgeClient.Chat(ctx, messages)
		if err == nil {
			judgeParser.Feed(full)
			judgeParser.Finalize()
			result.VerdictOneLiner = judgeParser.oneLiner
			result.VerdictFullBody = judgeParser.fullBody
		}
		if err != nil {
			return nil, fmt.Errorf("adjudicator: %w", err)
		}
	}

	// Generate Report
	if err := e.saveReport(result); err != nil {
		// Log error but don't fail the debate?
		fmt.Printf("Warning: Failed to save report: %v\n", err)
	}

	return result, nil
}

func (e *Executor) saveReport(r *Result) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("reports/debate_%s.md", timestamp)

	// Ensure dir exists
	if err := os.MkdirAll("reports", 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write Content
	tmpl := `# Debate Report
> Generated by Dialecta at %s

## 💡 Pro One-Liner
%s

## 🟢 Affirmative Argument (Full)
%s

---

## 💡 Con One-Liner
%s

## 🔴 Negative Argument (Full)
%s

---

## 💡 Verdict
%s

## ⚖️ Full Adjudication
%s
`
	content := fmt.Sprintf(tmpl,
		time.Now().Format(time.RFC1123),
		r.ProOneLiner, r.ProFullBody,
		r.ConOneLiner, r.ConFullBody,
		r.VerdictOneLiner, r.VerdictFullBody,
	)

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	r.ReportPath = filename
	return nil
}
