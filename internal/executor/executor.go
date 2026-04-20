package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wechat-task/api/internal/llm"
	"github.com/wechat-task/api/internal/logger"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
	"github.com/wechat-task/api/internal/service"
)

// Executor handles skill execution logic.
type Executor struct {
	subscriptionRepo  *repository.SkillSubscriptionRepository
	skillRepo         *repository.SkillRepository
	executionLogRepo  *repository.SkillExecutionLogRepository
	userLLMConfigRepo *repository.UserLLMConfigRepository
	channelService    *service.ChannelService
	maxConcurrency    int
	executionTimeout  time.Duration
}

// NewExecutor creates a new Executor.
func NewExecutor(
	subscriptionRepo *repository.SkillSubscriptionRepository,
	skillRepo *repository.SkillRepository,
	executionLogRepo *repository.SkillExecutionLogRepository,
	userLLMConfigRepo *repository.UserLLMConfigRepository,
	channelService *service.ChannelService,
	maxConcurrency int,
	executionTimeout time.Duration,
) *Executor {
	return &Executor{
		subscriptionRepo:  subscriptionRepo,
		skillRepo:         skillRepo,
		executionLogRepo:  executionLogRepo,
		userLLMConfigRepo: userLLMConfigRepo,
		channelService:    channelService,
		maxConcurrency:    maxConcurrency,
		executionTimeout:  executionTimeout,
	}
}

// RunBatch fetches and executes all due subscriptions up to batchSize.
func (e *Executor) RunBatch(ctx context.Context, batchSize int) {
	subs, err := e.subscriptionRepo.GetSubscriptionsDueForExecution(batchSize)
	if err != nil {
		logger.Errorf("Failed to fetch due subscriptions: %v", err)
		return
	}
	if len(subs) == 0 {
		return
	}

	logger.Infof("Executing batch of %d subscription(s)", len(subs))

	sem := make(chan struct{}, e.maxConcurrency)
	var wg sync.WaitGroup

	for _, sub := range subs {
		wg.Add(1)
		sem <- struct{}{}
		go func(s model.SkillSubscription) {
			defer wg.Done()
			defer func() { <-sem }()
			e.Execute(ctx, s)
		}(sub)
	}

	wg.Wait()
}

// Execute runs a single skill subscription.
func (e *Executor) Execute(ctx context.Context, sub model.SkillSubscription) {
	startTime := time.Now()

	// Load skill
	skill, err := e.skillRepo.GetByID(sub.SkillID)
	if err != nil {
		logger.Errorf("Subscription %d: failed to load skill %d: %v", sub.ID, sub.SkillID, err)
		e.scheduleNextRun(&sub)
		return
	}

	// Skip if skill is not published
	if skill.Status != model.SkillStatusPublished {
		logger.Warnf("Subscription %d: skill %d is %s, skipping", sub.ID, sub.SkillID, skill.Status)
		e.scheduleNextRun(&sub)
		return
	}

	// Resolve LLM config
	provider, llmReq, resolveErr := e.resolveLLM(ctx, &sub, skill)
	if resolveErr != nil {
		e.recordLog(sub.ID, "failed", "", "", resolveErr.Error(), 0, time.Since(startTime).Milliseconds())
		e.scheduleNextRun(&sub)
		return
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.executionTimeout)
	defer cancel()

	resp, execErr := provider.Complete(execCtx, *llmReq)
	durationMs := time.Since(startTime).Milliseconds()

	if execErr != nil {
		errMsg := execErr.Error()
		e.recordLog(sub.ID, "failed", llmReq.Prompt, "", errMsg, 0, durationMs)
		logger.Errorf("Subscription %d: execution failed: %v", sub.ID, execErr)
	} else {
		tokenUsage := resp.TokenUsage.Total
		e.recordLog(sub.ID, "success", llmReq.Prompt, resp.Content, "", tokenUsage, durationMs)

		// Increment execution count
		if err := e.skillRepo.IncrementExecutionCount(skill.ID); err != nil {
			logger.Errorf("Subscription %d: failed to increment execution count: %v", sub.ID, err)
		}

		// Deliver to channel if configured
		if sub.ChannelID != nil && sub.BotID != nil {
			if err := e.channelService.DeliverToChannel(ctx, *sub.BotID, *sub.ChannelID, resp.Content); err != nil {
				logger.Warnf("Subscription %d: delivery failed: %v", sub.ID, err)
			}
		}

		logger.Infof("Subscription %d: executed successfully (tokens=%d, duration=%dms)", sub.ID, tokenUsage, durationMs)
	}

	e.scheduleNextRun(&sub)
}

// resolveLLM resolves the LLM config and creates the provider + request.
func (e *Executor) resolveLLM(ctx context.Context, sub *model.SkillSubscription, skill *model.Skill) (llm.LLMProvider, *llm.LLMRequest, error) {
	var providerName, apiKey, modelStr, baseURL string
	var temperature float64
	var maxTokens int

	// Check subscription-level LLM config first
	if sub.Config.LLMConfig != nil && sub.Config.LLMConfig.APIKey != "" {
		cfg := sub.Config.LLMConfig
		providerName = cfg.Provider
		apiKey = cfg.APIKey
		modelStr = cfg.Model
		baseURL = cfg.BaseURL
		temperature = cfg.Temperature
		maxTokens = cfg.MaxTokens
	} else {
		// Fall back to user's default LLM config
		userCfg, err := e.userLLMConfigRepo.GetDefaultByUserID(sub.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("no LLM config found for user %d: %w", sub.UserID, err)
		}
		providerName = userCfg.Provider
		apiKey = userCfg.APIKey
		modelStr = userCfg.Model
		if userCfg.BaseURL != nil {
			baseURL = *userCfg.BaseURL
		}
	}

	// Create provider
	var provider llm.LLMProvider
	switch providerName {
	case "anthropic":
		provider = &llm.AnthropicProvider{APIKey: apiKey, BaseURL: baseURL}
	case "openai":
		provider = &llm.OpenAIProvider{APIKey: apiKey, BaseURL: baseURL}
	default:
		return nil, nil, fmt.Errorf("unsupported LLM provider: %s", providerName)
	}

	// Build prompt
	prompt := BuildPrompt(skill, sub.Config.Parameters)

	req := &llm.LLMRequest{
		Model:        modelStr,
		Prompt:       prompt.User,
		SystemPrompt: prompt.System,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
	}

	return provider, req, nil
}

// recordLog creates an execution log entry.
func (e *Executor) recordLog(subscriptionID uint, status, input, output, errMsg string, tokenUsage int, durationMs int64) {
	now := time.Now()
	logEntry := &model.SkillExecutionLog{
		SubscriptionID: subscriptionID,
		Status:         status,
		Input:          input,
		Output:         output,
		TokenUsage:     tokenUsage,
		DurationMs:     durationMs,
		ExecutedAt:     now,
		CompletedAt:    now,
	}
	if errMsg != "" {
		logEntry.Error = &errMsg
	}
	if err := e.executionLogRepo.Create(logEntry); err != nil {
		logger.Errorf("Failed to create execution log for subscription %d: %v", subscriptionID, err)
	}
}

// scheduleNextRun calculates and updates the next run time for a subscription.
func (e *Executor) scheduleNextRun(sub *model.SkillSubscription) {
	if sub.ScheduleCron == nil || *sub.ScheduleCron == "" {
		// No schedule, clear NextRunAt
		e.subscriptionRepo.UpdateNextRunAt(sub.ID, nil)
		return
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(*sub.ScheduleCron)
	if err != nil {
		logger.Errorf("Subscription %d: invalid cron %q: %v", sub.ID, *sub.ScheduleCron, err)
		e.subscriptionRepo.UpdateNextRunAt(sub.ID, nil)
		return
	}

	// Load timezone
	loc := time.UTC
	if sub.TimeZone != "" {
		if l, err := time.LoadLocation(sub.TimeZone); err == nil {
			loc = l
		}
	}

	nextRun := schedule.Next(time.Now().In(loc))
	if err := e.subscriptionRepo.UpdateNextRunAt(sub.ID, &nextRun); err != nil {
		logger.Errorf("Subscription %d: failed to update NextRunAt: %v", sub.ID, err)
	}
}
