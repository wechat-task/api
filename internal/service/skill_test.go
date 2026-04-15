package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateSkillContent_EmptyContent tests that empty content returns error
func TestValidateSkillContent_EmptyContent(t *testing.T) {
	err := ValidateSkillContent("")
	assert.Error(t, err, "Empty content should return error")
	assert.Contains(t, err.Error(), "empty", "Error should mention empty content")
}

// TestValidateSkillContent_MinimalValidContent tests that minimal valid content passes
func TestValidateSkillContent_MinimalValidContent(t *testing.T) {
	// Simplest Anthropic skill markdown format
	minimalContent := `# My Skill

This is a simple skill description.

## Parameters
- param1: string`

	err := ValidateSkillContent(minimalContent)
	assert.NoError(t, err, "Minimal valid content should pass validation")
}

// TestValidateSkillContent_MissingTitle tests that missing title returns error
func TestValidateSkillContent_MissingTitle(t *testing.T) {
	content := `This is a skill without a title.

## Parameters
- param1: string`

	err := ValidateSkillContent(content)
	assert.Error(t, err, "Content without title should return error")
	assert.Contains(t, err.Error(), "title", "Error should mention missing title")
}

// TestValidateSkillContent_ValidAnthropicFormat tests that valid Anthropic format passes
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

// TestValidateSkillContent_OnlyWhitespace tests that whitespace-only content returns error
func TestValidateSkillContent_OnlyWhitespace(t *testing.T) {
	content := "   \n\t\n  "
	err := ValidateSkillContent(content)
	assert.Error(t, err, "Whitespace-only content should return error")
}

// TestValidateSkillContent_TitleWithExtraHashtags tests that title with extra hashtags passes
func TestValidateSkillContent_TitleWithExtraHashtags(t *testing.T) {
	content := `# My Skill #awesome

Description here.`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Title with extra text after # should pass")
}

// TestValidateSkillContent_MultipleTitles tests that multiple titles passes (checks at least one)
func TestValidateSkillContent_MultipleTitles(t *testing.T) {
	content := `# Main Title

Some description.

## Subtitle

More content.`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Multiple title levels should pass")
}

// TestValidateSkillContent_MissingParametersSection tests that missing parameters section passes (not required in phase 1)
func TestValidateSkillContent_MissingParametersSection(t *testing.T) {
	content := `# Simple Skill

This skill has no parameters section, which should be allowed in phase 1.`

	err := ValidateSkillContent(content)
	assert.NoError(t, err, "Missing parameters section should be allowed in phase 1")
}

// TestValidateSkillContent_MaxLength tests content length limit
func TestValidateSkillContent_MaxLength(t *testing.T) {
	// Create long content (>100KB)
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
