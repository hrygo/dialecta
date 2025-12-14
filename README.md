# Dialecta - 多角色辩论系统

基于 Go 的 CLI 工具，实现 **Multi-Persona Debate** 工作流，打破单一 LLM 的幻觉和盲目顺从。

```
┌─────────────┐     ┌─────────────┐
│  🟢 正方    │     │  🔴 反方    │
│ Affirmative │ ──▶ │  Negative   │
└──────┬──────┘     └──────┬──────┘
       │     并行执行      │
       └────────┬──────────┘
                ▼
       ┌─────────────────┐
       │   ⚖️ 裁决方     │
       │   Adjudicator   │
       └─────────────────┘
```

## 快速开始

```bash
# 设置 API Key（按需设置）
export DEEPSEEK_API_KEY="your-deepseek-key"
export GEMINI_API_KEY="your-gemini-key"
export DASHSCOPE_API_KEY="your-dashscope-key"

# 构建
go build -o dialecta ./cmd/dialecta

# 使用
./dialecta proposal.md
echo "我们应该启动 AI 创业项目" | ./dialecta -
./dialecta --interactive
```

## 支持的提供商

| 提供商    | 环境变量                            | 默认模型           |
| --------- | ----------------------------------- | ------------------ |
| DeepSeek  | `DEEPSEEK_API_KEY`                  | `deepseek-chat`    |
| Gemini    | `GEMINI_API_KEY` / `GOOGLE_API_KEY` | `gemini-2.0-flash` |
| DashScope | `DASHSCOPE_API_KEY`                 | `qwen-plus`        |

## 命令行选项

```bash
dialecta [options] <file>

Options:
  --pro-provider     正方提供商 (default: deepseek)
  --pro-model        正方模型
  --con-provider     反方提供商 (default: deepseek)
  --con-model        反方模型
  --judge-provider   裁决提供商 (default: gemini)
  --judge-model      裁决模型
  --stream           流式输出 (default: true)
  --interactive      交互模式
```

## 示例

```bash
# 全部使用 DeepSeek
dialecta --judge-provider deepseek proposal.md

# 裁决使用 DeepSeek Reasoner
dialecta --judge-provider deepseek --judge-model deepseek-reasoner proposal.md

# 使用 DashScope (Qwen)
dialecta --pro-provider dashscope --con-provider dashscope --judge-provider dashscope proposal.md
```

## License

MIT
