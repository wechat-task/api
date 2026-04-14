# Skill System Design

**Date:** 2026-04-13  
**Author:** Claude Code  
**Status:** Draft

## 1. Overview

A skill system that allows users to create, share, subscribe to, and execute AI skills similar to Anthropic's AI agent skills. Skills can be scheduled to run automatically and results are delivered to users via messaging channels.

### 1.1 Core Requirements

1. **User-Centric Skill Creation**: All authenticated users can create their own skills
2. **Skill Sharing**: Users can make skills public for others to subscribe to
3. **Scheduling**: Skills can be scheduled to run at specific times using cron expressions
4. **Channel Delivery**: Results are delivered to users via existing messaging channels (WeChat, Lark)
5. **LLM Flexibility**: Free skills use system LLM, paid/special skills require user-provided LLM APIs
6. **Anthropic Skill Format**: Skills follow Anthropic's markdown format specification

## 2. Architecture Design

### 2.1 Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Skill System                             │
├──────────────┬──────────────┬──────────────┬───────────────┤
│   Web API    │   Scheduler  │   Executor   │   Publisher   │
│   (Gin)      │   (Cron)     │   (LLM)      │   (Channels)  │
└──────────────┴──────────────┴──────────────┴───────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                     Data Layer                              │
├──────────────┬──────────────┬──────────────┬───────────────┤
│   Skill      │  Subscription│   Execution  │   User LLM    │
│   Repository │  Repository  │   Log        │   Config      │
└──────────────┴──────────────┴──────────────┴───────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                     Storage                                 │
│                     PostgreSQL                              │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Data Models

#### 2.2.1 Skill
```go
type Skill struct {
    ID          uint           // Primary key
    UserID      uint           // Creator ID
    Name        string         // Skill name
    Description string         // Skill description
    Version     string         // Semantic version
    Content     string         // Markdown content (Anthropic format)
    Visibility  SkillVisibility // private/public/unlisted
    Status      SkillStatus     // draft/published/archived
    Category    string          // Category for organization
    Tags        []string        // Search tags
    
    // Pricing and LLM
    IsFree      bool           // Free to use
    UsesSystemLLM bool         // Uses system LLM or requires user LLM
    MaxTokens   int            // Max tokens for execution
    
    // Parameters
    Parameters  SkillParameters // Parameter definitions
    
    // Scheduling
    ScheduleCron *string       // Default cron schedule
    
    // Statistics
    SubscriberCount int        // Number of subscribers
    ExecutionCount  int        // Number of executions
    
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### 2.2.2 SkillSubscription
```go
type SkillSubscription struct {
    ID         uint                 // Primary key
    UserID     uint                 // Subscriber ID
    SkillID    uint                 // Skill being subscribed to
    
    // Execution configuration
    Config     SkillExecutionConfig // Parameter values and LLM config
    Status     string               // active/paused/cancelled
    
    // Delivery channel
    BotID      *uint               // Optional: which bot to use
    ChannelID  *uint               // Optional: which channel to use
    
    // Schedule overrides
    ScheduleCron *string           // Override skill's default schedule
    TimeZone     string            // Timezone for scheduling
    
    // Next execution time (for scheduler)
    NextRunAt  *time.Time
    
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

#### 2.2.3 SkillExecutionLog
```go
type SkillExecutionLog struct {
    ID              uint      // Primary key
    SubscriptionID  uint      // Related subscription
    Status          string    // success/failed/cancelled
    Input           string    // Input sent to LLM
    Output          string    // Output from LLM
    Error           *string   // Error message if failed
    TokenUsage      int       // Tokens used
    DurationMs      int64     // Execution duration
    ExecutedAt      time.Time // When execution started
    CompletedAt     time.Time // When execution completed
}
```

#### 2.2.4 UserLLMConfig
```go
type UserLLMConfig struct {
    ID         uint      // Primary key
    UserID     uint      // User who owns this config
    Name       string    // Configuration name (e.g., "My OpenAI")
    Provider   string    // openai/anthropic/azure/local
    APIKey     string    // Encrypted API key
    Model      string    // Model name
    BaseURL    *string   // Custom base URL (for self-hosted)
    IsDefault  bool      // Default configuration for user
    
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### 2.3 API Endpoints

#### Skill Management
- `POST /api/v1/skills` - Create a new skill
- `GET /api/v1/skills` - List skills (with filters: mine, public, category)
- `GET /api/v1/skills/:id` - Get skill details
- `PUT /api/v1/skills/:id` - Update skill
- `DELETE /api/v1/skills/:id` - Delete skill
- `POST /api/v1/skills/:id/publish` - Publish skill
- `POST /api/v1/skills/:id/archive` - Archive skill

#### Subscription Management
- `POST /api/v1/skills/:id/subscribe` - Subscribe to a skill
- `GET /api/v1/skills/subscriptions` - List my subscriptions
- `PUT /api/v1/skills/subscriptions/:id` - Update subscription config
- `DELETE /api/v1/skills/subscriptions/:id` - Unsubscribe
- `POST /api/v1/skills/subscriptions/:id/pause` - Pause subscription
- `POST /api/v1/skills/subscriptions/:id/resume` - Resume subscription

#### User LLM Configuration
- `POST /api/v1/llm-configs` - Add LLM configuration
- `GET /api/v1/llm-configs` - List my LLM configurations
- `PUT /api/v1/llm-configs/:id` - Update LLM configuration
- `DELETE /api/v1/llm-configs/:id` - Delete LLM configuration
- `POST /api/v1/llm-configs/:id/set-default` - Set as default

#### Execution Management
- `POST /api/v1/skills/subscriptions/:id/execute` - Manual execution
- `GET /api/v1/skills/executions` - List execution history
- `GET /api/v1/skills/executions/:id` - Get execution details

## 3. Implementation Phases

### Phase 1: Foundation (Week 1)
**Goal**: Basic skill creation and storage
- [ ] Data models and migrations
- [ ] Repository layer for skills and subscriptions
- [ ] Basic API endpoints for skill CRUD
- [ ] Authentication and authorization checks
- [ ] Markdown content validation

### Phase 2: Subscription System (Week 2)
**Goal**: Users can subscribe to skills
- [ ] Subscription management APIs
- [ ] Parameter validation for subscriptions
- [ ] Basic skill listing with filters
- [ ] Visibility and permission controls

### Phase 3: LLM Integration (Week 3)
**Goal**: Skill execution with LLMs
- [ ] User LLM configuration storage
- [ ] LLM client abstraction (OpenAI, Anthropic, Azure)
- [ ] Skill content parsing and template rendering
- [ ] Parameter substitution in skill templates
- [ ] Basic manual execution API

### Phase 4: Scheduling System (Week 4)
**Goal**: Automated scheduling of skills
- [ ] Cron expression parsing and validation
- [ ] Scheduler service that runs in background
- [ ] Next execution time calculation
- [ ] Subscription status management (active/paused)
- [ ] Execution queue and worker pool

### Phase 5: Channel Delivery (Week 5)
**Goal**: Deliver results to messaging channels
- [ ] Integration with existing ChannelService
- [ ] Message formatting and delivery
- [ ] Execution logging and error handling
- [ ] Retry mechanism for failed deliveries

### Phase 6: Advanced Features (Week 6+)
**Goal**: Polish and advanced capabilities
- [ ] Skill import/export (Anthropic format)
- [ ] Skill versioning
- [ ] Webhook triggers (in addition to cron)
- [ ] Execution statistics and analytics
- [ ] Rate limiting and usage quotas
- [ ] Skill discovery and search

## 4. Technical Decisions

### 4.1 LLM Provider Abstraction
Create a generic `LLMProvider` interface that supports:
- OpenAI (GPT models)
- Anthropic (Claude models)
- Azure OpenAI
- Local models (via Ollama/OpenAI-compatible APIs)

### 4.2 Scheduler Design
- Use `robfig/cron` for cron expression parsing
- Store `next_run_at` in database for efficient querying
- Run scheduler as a goroutine in main application
- Use database locks to prevent duplicate execution

### 4.3 Template Rendering
- Parse markdown content to extract parameters
- Use Go's `text/template` for parameter substitution
- Support both simple `{{.param}}` and advanced template logic
- Validate template syntax on skill creation

### 4.4 Security Considerations
- Encrypt user LLM API keys at rest
- Validate cron expressions to prevent injection
- Rate limit skill executions per user
- Sanitize markdown content to prevent XSS
- Audit log for all skill executions

### 4.5 Error Handling
- Retry failed LLM calls with exponential backoff
- Store execution logs for debugging
- Notify users of failed executions via channel
- Graceful degradation when LLM services are unavailable

## 5. Database Schema

```sql
-- Skills table
CREATE TABLE skills (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) DEFAULT '1.0.0',
    content TEXT NOT NULL,
    visibility VARCHAR(20) DEFAULT 'private',
    status VARCHAR(20) DEFAULT 'draft',
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    is_free BOOLEAN DEFAULT true,
    uses_system_llm BOOLEAN DEFAULT true,
    max_tokens INTEGER DEFAULT 2000,
    parameters JSONB DEFAULT '{}',
    schedule_cron VARCHAR(100),
    subscriber_count INTEGER DEFAULT 0,
    execution_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Skill subscriptions
CREATE TABLE skill_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    skill_id INTEGER NOT NULL REFERENCES skills(id),
    config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    bot_id INTEGER REFERENCES bots(id),
    channel_id INTEGER REFERENCES channels(id),
    schedule_cron VARCHAR(100),
    time_zone VARCHAR(50) DEFAULT 'UTC',
    next_run_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, skill_id)
);

-- User LLM configurations
CREATE TABLE user_llm_configs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    api_key TEXT NOT NULL, -- encrypted
    model VARCHAR(100) NOT NULL,
    base_url VARCHAR(255),
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Execution logs
CREATE TABLE skill_execution_logs (
    id SERIAL PRIMARY KEY,
    subscription_id INTEGER NOT NULL REFERENCES skill_subscriptions(id),
    status VARCHAR(20) NOT NULL,
    input TEXT,
    output TEXT,
    error TEXT,
    token_usage INTEGER DEFAULT 0,
    duration_ms BIGINT DEFAULT 0,
    executed_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NOT NULL
);
```

## 6. Success Metrics

1. **User Adoption**: Number of skills created and subscribed to
2. **Execution Reliability**: Percentage of successful executions
3. **Performance**: Average execution time under 30 seconds
4. **Scalability**: Support for 1000+ concurrent subscriptions
5. **User Satisfaction**: Low unsubscribe rate and high engagement

## 7. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| LLM API costs | High | Implement usage quotas and rate limiting |
| Schedule accuracy | Medium | Use robust cron library with timezone support |
| Skill content security | High | Sanitize markdown and validate templates |
| Database performance | Medium | Index critical fields and archive old logs |
| Channel delivery failures | Medium | Implement retry queue with exponential backoff |

## 8. Open Questions

1. Should skills support versioning for breaking changes?
2. How to handle skill deprecation and migration?
3. What notification system for execution failures?
4. How to monetize premium skills?
5. Should there be a skill marketplace/review system?

---

*This design is ready for implementation. The phased approach allows for incremental delivery while managing complexity.*