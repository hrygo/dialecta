package debate

import (
	"context"
	"fmt"
	"sync"

	"github.com/huangzhonghui/dialecta/internal/config"
	"github.com/huangzhonghui/dialecta/internal/llm"
	"github.com/huangzhonghui/dialecta/internal/prompt"
)

// Result holds the complete debate result
type Result struct {
	Material    string // 原始材料
	ProArgument string // 正方论述
	ConArgument string // 反方论述
	Verdict     string // 裁决报告
}

// Executor orchestrates the debate process
type Executor struct {
	cfg     *config.Config
	stream  bool
	onPro   func(string) // 正方流式回调
	onCon   func(string) // 反方流式回调
	onJudge func(string) // 裁决流式回调
}

// NewExecutor creates a new debate executor
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		cfg:    cfg,
		stream: false,
	}
}

// SetStream enables streaming mode with callbacks
func (e *Executor) SetStream(onPro, onCon, onJudge func(string)) {
	e.stream = true
	e.onPro = onPro
	e.onCon = onCon
	e.onJudge = onJudge
}

// Execute runs the full debate workflow
func (e *Executor) Execute(ctx context.Context, material string) (*Result, error) {
	result := &Result{Material: material}

	// Phase 1: 并行执行正反方辩论
	var wg sync.WaitGroup
	var proErr, conErr error

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
			result.ProArgument, err = client.ChatStream(ctx, messages, e.onPro)
		} else {
			result.ProArgument, err = client.Chat(ctx, messages)
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
			result.ConArgument, err = client.ChatStream(ctx, messages, e.onCon)
		} else {
			result.ConArgument, err = client.Chat(ctx, messages)
		}
		if err != nil {
			conErr = fmt.Errorf("negative: %w", err)
		}
	}()

	wg.Wait()

	// 检查错误
	if proErr != nil {
		return nil, proErr
	}
	if conErr != nil {
		return nil, conErr
	}

	// Phase 2: 裁决
	judgeClient, err := llm.NewClient(e.cfg.JudgeRole.ToLLMConfig())
	if err != nil {
		return nil, fmt.Errorf("create judge client: %w", err)
	}

	messages := prompt.BuildAdjudicatorMessages(material, result.ProArgument, result.ConArgument)
	if e.stream && e.onJudge != nil {
		result.Verdict, err = judgeClient.ChatStream(ctx, messages, e.onJudge)
	} else {
		result.Verdict, err = judgeClient.Chat(ctx, messages)
	}
	if err != nil {
		return nil, fmt.Errorf("adjudicator: %w", err)
	}

	return result, nil
}
