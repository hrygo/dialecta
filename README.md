# Dialecta - 多角色辩论系统

一个基于 Go 的 CLI 工具，实现 **Multi-Persona Debate** 工作流，打破单一 LLM 的幻觉和盲目顺从。

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
# 设置 API Key
export OPENROUTER_API_KEY="your-key-here"

# 构建
go build -o dialecta ./cmd/dialecta

# 分析文件
./dialecta proposal.md

# 管道输入
echo "我们应该在明年启动一个 AI 创业项目" | ./dialecta -

# 交互模式
./dialecta --interactive
```

## 命令行选项

```bash
dialecta [options] <file>

Options:
  --pro-model     正方模型 (default: deepseek/deepseek-chat)
  --con-model     反方模型 (default: deepseek/deepseek-chat)
  --judge-model   裁决模型 (default: anthropic/claude-sonnet-4-20250514)
  --stream        流式输出 (default: true)
  --interactive   交互模式
```

## 配置

### 环境变量

| 变量名                | 说明               | 默认值                         |
| --------------------- | ------------------ | ------------------------------ |
| `OPENROUTER_API_KEY`  | OpenRouter API Key | -                              |
| `OPENAI_API_KEY`      | 备选 API Key       | -                              |
| `OPENROUTER_BASE_URL` | API 基础 URL       | `https://openrouter.ai/api/v1` |

### 模型策略

- **正方/反方**: 使用快速模型，Temperature 0.8 激发发散思维
- **裁决方**: 使用强逻辑模型，Temperature 0.1 保持理性收敛

## 输出示例

```
╔══════════════════════════════════════════════════════════════╗
║           🎭 Dialecta - 多角色辩论系统                        ║
╚══════════════════════════════════════════════════════════════╝

🟢 正方论述 (The Affirmative)
────────────────────────────────────────────────────────────────
【正方核心立场】：...
【关键支撑论据】：...

🔴 反方论述 (The Negative)
────────────────────────────────────────────────────────────────
【反方核心驳斥】：...
【关键风险/漏洞】：...

⚖️ 裁决方报告 (The Adjudicator)
────────────────────────────────────────────────────────────────
## ⚖️ 综合裁决报告
### 1. 争议焦点分析
...
### 3. 最终裁决
* **综合评分**：XX / 100
```

## License

MIT
