package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wechat-task/api/internal/model"
)

// Prompt holds the system and user prompts for an LLM call.
type Prompt struct {
	System string
	User   string
}

// BuildPrompt constructs system and user prompts from a skill and subscription parameters.
func BuildPrompt(skill *model.Skill, params map[string]any) Prompt {
	system := fmt.Sprintf("You are executing a skill named %q. Follow the instructions below.", skill.Name)

	user := skill.Content
	if len(params) > 0 {
		var sb strings.Builder
		sb.WriteString("\n\n## Provided Parameters\n")
		// Sort keys for deterministic output
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, params[k]))
		}
		user += sb.String()
	}

	return Prompt{System: system, User: user}
}
