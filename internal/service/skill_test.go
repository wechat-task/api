package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateSkillContent_EmptyContent 测试空内容应该返回错误
func TestValidateSkillContent_EmptyContent(t *testing.T) {
	err := ValidateSkillContent("")
	assert.Error(t, err, "Empty content should return error")
	assert.Contains(t, err.Error(), "empty", "Error should mention empty content")
}

// TestValidateSkillContent_MinimalValidContent 测试最小有效内容应该通过
func TestValidateSkillContent_MinimalValidContent(t *testing.T) {
	// 最简单的Anthropic技能markdown格式
	minimalContent := `# My Skill

This is a simple skill description.

## Parameters
- param1: string`

	err := ValidateSkillContent(minimalContent)
	assert.NoError(t, err, "Minimal valid content should pass validation")
}

// TestValidateSkillContent_MissingTitle 测试缺少标题应该返回错误
func TestValidateSkillContent_MissingTitle(t *testing.T) {
	content := `This is a skill without a title.

## Parameters
- param1: string`

	err := ValidateSkillContent(content)
	assert.Error(t, err, "Content without title should return error")
	assert.Contains(t, err.Error(), "title", "Error should mention missing title")
}

// TestValidateSkillContent_ValidAnthropicFormat 测试有效的Anthropic格式应该通过
func TestValidateSkillContent_ValidAnthropicFormat(t *testing.T) {
	content := `# Weather Assistant

Get current weather information for a location.

## Parameters
- location: string - The city name or coordinates
- unit: string - Temperature unit (celsius, fahrenheit), default: celsius

## Examples
### Example 1
User: What's the weather in Beijing?
Assistant: The weather in Beijing is sunny with 25°C.

### Example 2
User: Weather in New York in fahrenheit
Assistant: The weather in New York is cloudy with 68°F.

## Instructions
1. Extract location and unit from the request
2. Call weather API with the parameters
3. Format the response in a friendly way`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Valid Anthropic format should pass validation")
}

// TestValidateSkillContent_OnlyWhitespace 测试只有空白字符的内容应该返回错误
func TestValidateSkillContent_OnlyWhitespace(t *testing.T) {
	content := "   \n\t\n  "
	err := ValidateSkillContent(content)
	assert.Error(t, err, "Whitespace-only content should return error")
}

// TestValidateSkillContent_TitleWithExtraHashtags 测试标题有多个#号应该通过
func TestValidateSkillContent_TitleWithExtraHashtags(t *testing.T) {
	content := `# My Skill #awesome

Description here.`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Title with extra text after # should pass")
}

// TestValidateSkillContent_MultipleTitles 测试多个标题应该通过（只检查至少一个）
func TestValidateSkillContent_MultipleTitles(t *testing.T) {
	content := `# Main Title

Some description.

## Subtitle

More content.`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Multiple title levels should pass")
}

// TestValidateSkillContent_MissingParametersSection 测试缺少参数部分应该通过（第一阶段不强制要求）
func TestValidateSkillContent_MissingParametersSection(t *testing.T) {
	content := `# Simple Skill

This skill has no parameters section, which should be allowed in phase 1.`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Missing parameters section should be allowed in phase 1")
}

// TestValidateSkillContent_MaxLength 测试内容长度限制
func TestValidateSkillContent_MaxLength(t *testing.T) {
	// 创建超长内容（超过100KB）
	var sb strings.Builder
	sb.WriteString("# Very Long Skill\n\n")
	for i := range 50000 {
		sb.WriteString("Line ")
		sb.WriteByte(byte('A' + (i % 26)))
		sb.WriteString("\n")
	}
	longContent := sb.String()

	err := ValidateSkillContent(longContent)
	assert.NoError(t, err, "Long content should pass (no length limit in phase 1)")
}
