package prompt

import (
	"fmt"

	"github.com/hrygo/dialecta/internal/llm"
)

// BuildAffirmativeMessages builds the messages for the Affirmative model
func BuildAffirmativeMessages(material string) []llm.Message {
	return []llm.Message{
		{Role: "system", Content: AffirmativeSystemPrompt},
		{Role: "user", Content: fmt.Sprintf("**用户提供的材料如下：**\n\n%s", material)},
	}
}

// BuildNegativeMessages builds the messages for the Negative model
func BuildNegativeMessages(material string) []llm.Message {
	return []llm.Message{
		{Role: "system", Content: NegativeSystemPrompt},
		{Role: "user", Content: fmt.Sprintf("**用户提供的材料如下：**\n\n%s", material)},
	}
}

// BuildAdjudicatorMessages builds the messages for the Adjudicator model
func BuildAdjudicatorMessages(material, proArgument, conArgument string) []llm.Message {
	userContent := fmt.Sprintf(`**输入数据：**

**【原始材料】**：
%s

**【正方观点】**：
%s

**【反方观点】**：
%s`, material, proArgument, conArgument)

	return []llm.Message{
		{Role: "system", Content: AdjudicatorSystemPrompt},
		{Role: "user", Content: userContent},
	}
}
