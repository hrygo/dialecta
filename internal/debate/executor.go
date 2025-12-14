package debate

import (
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
	client  *llm.Client
	cfg     *config.Config
	stream  bool
	onPro   func(string) // 正方流式回调
	onCon   func(string) // 反方流式回调
	onJudge func(string) // 裁决流式回调
}

// NewExecutor creates a new debate executor
func NewExecutor(cfg *config.Config) *Executor {
	client := llm.NewClient(cfg.APIKey, cfg.BaseURL)
	return &Executor{
		client: client,
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
func (e *Executor) Execute(material string) (*Result, error) {
	result := &Result{Material: material}

	// Phase 1: 并行执行正反方辩论
	var wg sync.WaitGroup
	var proErr, conErr error

	wg.Add(2)

	// 正方
	go func() {
		defer wg.Done()
		messages := prompt.BuildAffirmativeMessages(material)
		mc := e.cfg.ProModel

		var err error
		if e.stream && e.onPro != nil {
			result.ProArgument, err = e.client.ChatStream(mc.Model, messages, mc.Temperature, mc.MaxTokens, e.onPro)
		} else {
			result.ProArgument, err = e.client.Chat(mc.Model, messages, mc.Temperature, mc.MaxTokens)
		}
		if err != nil {
			proErr = fmt.Errorf("affirmative: %w", err)
		}
	}()

	// 反方
	go func() {
		defer wg.Done()
		messages := prompt.BuildNegativeMessages(material)
		mc := e.cfg.ConModel

		var err error
		if e.stream && e.onCon != nil {
			result.ConArgument, err = e.client.ChatStream(mc.Model, messages, mc.Temperature, mc.MaxTokens, e.onCon)
		} else {
			result.ConArgument, err = e.client.Chat(mc.Model, messages, mc.Temperature, mc.MaxTokens)
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
	messages := prompt.BuildAdjudicatorMessages(material, result.ProArgument, result.ConArgument)
	mc := e.cfg.JudgeModel

	var err error
	if e.stream && e.onJudge != nil {
		result.Verdict, err = e.client.ChatStream(mc.Model, messages, mc.Temperature, mc.MaxTokens, e.onJudge)
	} else {
		result.Verdict, err = e.client.Chat(mc.Model, messages, mc.Temperature, mc.MaxTokens)
	}
	if err != nil {
		return nil, fmt.Errorf("adjudicator: %w", err)
	}

	return result, nil
}
