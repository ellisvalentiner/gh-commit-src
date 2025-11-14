package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

const MaxDiffLength = 30000 // set to 30k since large model has maximum context length is 32768 tokens.

// Default configuration values
const (
	defaultCommitPrompt = `You will examine and explain the given code changes and write a commit message in Conventional Commits format. 
		The first line of the commit message should be a 20 word Title summary include a type, optional scope, subject in text, seperated by a newline and the following body. 
		The types should be one of:
			- fix: for a bug fix
			- feat: for a new feature 
			- perf: for a performance improvement
			- revert: to revert a previous commit
		The body will explain the code change. Body will be formatted in well structured beautifully rendered and use relevant emojis
		if no code changes are detected, you will reply with no code change detected message.`
	defaultAzureAPIVersion     = "2024-12-01-preview"
	defaultCommitMessageSuffix = "Commit message as follows:"
)

// Default code block patterns for formatResponse
var defaultCodeBlockPatterns = []string{"```bash", "```plaintext", "```diff", "```python", "```javascript", "```go", "```java", "```csharp", "```ruby", "```php", "```html", "```css", "```json", "```xml", "```yaml", "```md", "```markdown", "```sql", "```shell", "```powershell", "```dockerfile", "```makefile", "```ini", "```apacheconf", "```nginx", "```git", "```vim", "```vimscrip", "```"}

func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

func getGitDiff() (string, error) {
	if !isGitRepository() {
		return "", fmt.Errorf("not a git repository")
	}
	cmd := exec.Command("git", "diff")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running git diff: %v", err)
	}
	diff := strings.TrimSpace(string(output))
	if diff == "" {
		cmd = exec.Command("git", "diff", "--staged")
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("error running git diff --staged: %v", err)
		}
		diff = strings.TrimSpace(string(output))
	}
	if diff == "" {
		return "", fmt.Errorf("no changes detected. Please stage or make changes before generating a commit message")
	}
	runes := []rune(diff)
	size := len(runes)
	if size > MaxDiffLength {
		runes = runes[:MaxDiffLength]
		return string(runes), fmt.Errorf("the total length was %d and only first 30k were used", size)
	}
	return diff, nil
}

func calculateTimeSaved(numCommits int, wordCount int) float64 {

	// Assuming an average typing speed of 40 words per minute
	wordsPerMinute := 40.0
	hoursSaved := float64(wordCount) / wordsPerMinute / 60
	return math.Round(hoursSaved*10) / 10
}

func getCommitStats() (int, int, error) {
	if !isGitRepository() {
		return 0, 0, fmt.Errorf("not a git repository")
	}
	gitCmd := exec.Command("git", "log", "--oneline")
	stdout, err := gitCmd.StdoutPipe()
	if err != nil {
		return 0, 0, err
	}
	if err := gitCmd.Start(); err != nil {
		return 0, 0, err
	}
	defer func() {
		if err := gitCmd.Wait(); err != nil {
			// Log error but don't fail if wc command succeeds
			_ = err
		}
	}()
	cmd := exec.Command("wc", "-lw")
	cmd.Stdin = stdout
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(string(output))
	numLines, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, err
	}
	numWords, err := strconv.Atoi(fields[1])
	if err != nil {
		return numLines, 0, err
	}
	return numLines, numWords, nil
}

func getCommitPrompt() string {
	if prompt := os.Getenv("PROMPT_OVERRIDE"); prompt != "" {
		return prompt
	}
	return defaultCommitPrompt
}

func getCommitMessageSuffix() string {
	if suffix := os.Getenv("COMMIT_MESSAGE_SUFFIX"); suffix != "" {
		return suffix
	}
	return defaultCommitMessageSuffix
}

func getDiffPrompt(diff string) []azopenai.ChatMessage {
	prompt := getCommitPrompt()
	suffix := getCommitMessageSuffix()
	messages := []azopenai.ChatMessage{
		{Role: to.Ptr(azopenai.ChatRoleSystem), Content: to.Ptr(prompt)},
		{Role: to.Ptr(azopenai.ChatRoleUser), Content: to.Ptr(diff)},
		{Role: to.Ptr(azopenai.ChatRoleSystem), Content: to.Ptr(suffix)},
	}
	return messages
}

func getPrompt(message string) []azopenai.ChatMessage {
	messages := []azopenai.ChatMessage{
		{Role: to.Ptr(azopenai.ChatRoleSystem), Content: to.Ptr(message)},
	}
	return messages
}

func getChatCompletionResponse(messages []azopenai.ChatMessage) (string, error) {
	err := godotenv.Load()
	if err != nil {
		// .env file is optional, so we ignore the error
		_ = err
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is not set. Please export OPENAI_API_KEY=<api_key>")
	}
	keyCredential, err := azopenai.NewKeyCredential(apiKey)
	if err != nil {
		return "", fmt.Errorf("error creating Azure OpenAI client: %v", err)
	}
	url := os.Getenv("OPENAI_URL")
	model := os.Getenv("OPENAI_MODEL")
	var client *azopenai.Client

	if strings.Contains(url, "azure") {
		apiVersion := getAzureAPIVersion()
		clientOptions := &azopenai.ClientOptions{
			ClientOptions: policy.ClientOptions{
				APIVersion: apiVersion,
			},
		}
		client, err = azopenai.NewClientWithKeyCredential(url, keyCredential, clientOptions)
		if err != nil {
			return "", fmt.Errorf("error creating Azure OpenAI client: %v", err)
		}
	} else {
		client, err = azopenai.NewClientForOpenAI(url, keyCredential, nil)
		if err != nil {
			return "", fmt.Errorf("error creating Azure OpenAI client: %v", err)
		}

	}
	if model == "" {
		model = openai.GPT4
	}

	options := azopenai.ChatCompletionsOptions{
		Messages:   messages,
		Deployment: model,
	}

	// Parse FINE_TUNE_PARAMS if provided
	fineTuneParams := os.Getenv("FINE_TUNE_PARAMS")
	if fineTuneParams != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(fineTuneParams), &params); err == nil {
			// Apply common parameters
			if temp, ok := params["temperature"].(float64); ok {
				options.Temperature = to.Ptr(float32(temp))
			}
			if maxTokens, ok := params["max_tokens"].(float64); ok {
				options.MaxTokens = to.Ptr(int32(maxTokens))
			}
			if topP, ok := params["top_p"].(float64); ok {
				options.TopP = to.Ptr(float32(topP))
			}
			if frequencyPenalty, ok := params["frequency_penalty"].(float64); ok {
				options.FrequencyPenalty = to.Ptr(float32(frequencyPenalty))
			}
			if presencePenalty, ok := params["presence_penalty"].(float64); ok {
				options.PresencePenalty = to.Ptr(float32(presencePenalty))
			}
		}
	}

	resp, err := client.GetChatCompletions(
		context.Background(),
		options,
		nil,
	)

	if err != nil {
		return "", fmt.Errorf("completion error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("api returned no choices in response")
	}

	if resp.Choices[0].Message.Content == nil {
		return "", fmt.Errorf("api returned empty content in response")
	}

	return *resp.Choices[0].Message.Content, nil
}

func getAzureAPIVersion() string {
	if version := os.Getenv("AZURE_API_VERSION"); version != "" {
		return version
	}
	return defaultAzureAPIVersion
}

func getCodeBlockPatterns() []string {
	if patternsEnv := os.Getenv("CODE_BLOCK_PATTERNS"); patternsEnv != "" {
		// Allow comma-separated list of patterns from environment
		patterns := strings.Split(patternsEnv, ",")
		// Trim whitespace from each pattern
		for i, pattern := range patterns {
			patterns[i] = strings.TrimSpace(pattern)
		}
		return patterns
	}
	return defaultCodeBlockPatterns
}

func formatResponse(response string) string {
	patterns := getCodeBlockPatterns()
	// Patterns ordered from most specific to least specific so language-specific markers are removed before generic ```
	for _, pattern := range patterns {
		response = strings.TrimPrefix(response, pattern)
		response = strings.TrimSuffix(response, pattern)
	}
	return response
}
