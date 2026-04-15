# Skill Status Flow Enforcement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enforce strict status-based access control for skill mutations — only draft skills can be edited, published skills are immutable, and archived skills cannot be re-published.

**Architecture:** Introduce two dedicated endpoints (`POST /skills/:id/publish`, `POST /skills/:id/archive`) to manage status transitions, and strip status changes from the generic update endpoint. The service layer enforces the state machine: `draft → published → archived` (terminal).

**Tech Stack:** Go, Gin, GORM, PostgreSQL

---

### Task 1: Add PublishSkill service method

**Files:**
- Modify: `internal/service/skill.go:209-270` (UpdateSkill method area)

**Step 1: Write the PublishSkill method**

Add after the `UpdateSkill` method in `internal/service/skill.go`:

```go
// PublishSkill publishes a draft skill
func (s *SkillService) PublishSkill(userID uint, skillID uint) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can publish")
	}

	if skill.Status == model.SkillStatusPublished {
		return nil, errors.New("skill is already published")
	}

	if skill.Status == model.SkillStatusArchived {
		return nil, errors.New("cannot publish archived skill")
	}

	// Only draft skills can be published
	if skill.Status != model.SkillStatusDraft {
		return nil, errors.New("only draft skills can be published")
	}

	skill.Status = model.SkillStatusPublished
	skill.UpdatedAt = time.Now()

	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}
```

**Step 2: Build to verify compilation**

Run: `go build -o server .`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/service/skill.go
git commit -m "feat: add PublishSkill service method"
```

---

### Task 2: Add ArchiveSkill service method

**Files:**
- Modify: `internal/service/skill.go`

**Step 1: Write the ArchiveSkill method**

Add after the `PublishSkill` method:

```go
// ArchiveSkill archives a published skill
func (s *SkillService) ArchiveSkill(userID uint, skillID uint) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can archive")
	}

	if skill.Status == model.SkillStatusArchived {
		return nil, errors.New("skill is already archived")
	}

	if skill.Status == model.SkillStatusDraft {
		return nil, errors.New("cannot archive draft skill, delete it instead")
	}

	if skill.Status != model.SkillStatusPublished {
		return nil, errors.New("only published skills can be archived")
	}

	skill.Status = model.SkillStatusArchived
	skill.UpdatedAt = time.Now()

	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}
```

**Step 2: Build to verify compilation**

Run: `go build -o server .`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/service/skill.go
git commit -m "feat: add ArchiveSkill service method"
```

---

### Task 3: Lock UpdateSkill to draft-only

**Files:**
- Modify: `internal/service/skill.go:209-270`

**Step 1: Rewrite UpdateSkill to reject non-draft skills and ignore status field**

Replace the existing `UpdateSkill` method with:

```go
// UpdateSkill updates a draft skill
func (s *SkillService) UpdateSkill(userID uint, skillID uint, updates model.Skill) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can update")
	}

	if skill.Status != model.SkillStatusDraft {
		return nil, errors.New("only draft skills can be modified")
	}

	// Validate content if updating
	if updates.Content != "" && updates.Content != skill.Content {
		if err := ValidateSkillContent(updates.Content); err != nil {
			return nil, err
		}
		skill.Content = updates.Content
	}

	if updates.Name != "" {
		skill.Name = updates.Name
	}
	if updates.Description != "" {
		skill.Description = updates.Description
	}
	if updates.Visibility != "" {
		skill.Visibility = updates.Visibility
	}
	if updates.Category != "" {
		skill.Category = updates.Category
	}
	if updates.Tags != nil {
		skill.Tags = updates.Tags
	}
	if updates.Parameters != nil {
		skill.Parameters = updates.Parameters
	}

	skill.UpdatedAt = time.Now()

	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}
```

Key changes from original:
- Removed the `updates.Status` field handling — status changes now go through dedicated Publish/Archive endpoints
- Replaced the partial "published skill" check with a single `skill.Status != SkillStatusDraft` guard
- Rejects all modifications to published AND archived skills

**Step 2: Build to verify compilation**

Run: `go build -o server .`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/service/skill.go
git commit -m "refactor: lock UpdateSkill to draft-only, remove status from updates"
```

---

### Task 4: Remove status from UpdateSkillRequest handler

**Files:**
- Modify: `internal/handler/skill.go:143-227`

**Step 1: Remove Status field from UpdateSkillRequest and its handling**

In `UpdateSkillRequest` struct, remove the `Status` field:

```go
type UpdateSkillRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Content     *string                `json:"content"`
	Visibility  *model.SkillVisibility `json:"visibility"`
	Category    *string                `json:"category"`
	Tags        *[]string              `json:"tags"`
	Parameters  *model.SkillParameters `json:"parameters"`
}
```

In the `UpdateSkill` handler, remove the status assignment block:

```go
// Remove these lines:
// if req.Status != nil {
//     updates.Status = *req.Status
// }
```

Also update the error handling in `UpdateSkill` to return `403` for the new draft-only error:

```go
skill, err := h.skillService.UpdateSkill(userID.(uint), uint(id), updates)
if err != nil {
    switch {
    case strings.HasPrefix(err.Error(), "unauthorized"):
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
    case strings.HasPrefix(err.Error(), "only draft skills"):
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }
    return
}
```

**Step 2: Add `strings` import if not already present**

**Step 3: Build to verify compilation**

Run: `go build -o server .`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/handler/skill.go
git commit -m "refactor: remove status from UpdateSkillRequest handler"
```

---

### Task 5: Add PublishSkill and ArchiveSkill handlers

**Files:**
- Modify: `internal/handler/skill.go`

**Step 1: Add PublishSkill handler**

```go
// PublishSkill godoc
// @Summary      Publish skill
// @Description  Publish a draft skill
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Skill ID"
// @Success      200  {object}  model.Skill  "Published skill"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      409  {object}  map[string]string  "状态冲突"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/{id}/publish [post]
func (h *SkillHandler) PublishSkill(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skill id"})
		return
	}

	skill, err := h.skillService.PublishSkill(userID.(uint), uint(id))
	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "unauthorized"):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "already published"),
			strings.Contains(err.Error(), "archived"),
			strings.Contains(err.Error(), "only draft"):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, skill)
}
```

**Step 2: Add ArchiveSkill handler**

```go
// ArchiveSkill godoc
// @Summary      Archive skill
// @Description  Archive a published skill
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Skill ID"
// @Success      200  {object}  model.Skill  "Archived skill"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      409  {object}  map[string]string  "状态冲突"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/{id}/archive [post]
func (h *SkillHandler) ArchiveSkill(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skill id"})
		return
	}

	skill, err := h.skillService.ArchiveSkill(userID.(uint), uint(id))
	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "unauthorized"):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "already archived"),
			strings.Contains(err.Error(), "draft"),
			strings.Contains(err.Error(), "only published"):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, skill)
}
```

**Step 3: Build to verify compilation**

Run: `go build -o server .`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/handler/skill.go
git commit -m "feat: add PublishSkill and ArchiveSkill handlers"
```

---

### Task 6: Register new routes

**Files:**
- Modify: `main.go:162-170`

**Step 1: Add publish and archive routes**

In the authenticated skill routes group, add two new routes:

```go
// Skill routes (authenticated)
skills := r.Group("/api/v1/skills")
skills.Use(middleware.Auth(jwtService))
{
    skills.POST("", skillHandler.CreateSkill)
    skills.GET("/me", skillHandler.GetMySkills)
    skills.GET("/search", skillHandler.SearchSkills)
    skills.POST("/:id/subscribe", skillHandler.SubscribeToSkill)
    skills.GET("/subscriptions", skillHandler.GetUserSubscriptions)
    skills.PUT("/:id", skillHandler.UpdateSkill)
    skills.DELETE("/:id", skillHandler.DeleteSkill)
    skills.POST("/:id/publish", skillHandler.PublishSkill)
    skills.POST("/:id/archive", skillHandler.ArchiveSkill)
}
```

Note: The existing `UpdateSkill` and `DeleteSkill` routes were not registered in the authenticated group — add them as well.

**Step 2: Build to verify compilation**

Run: `go build -o server .`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: register publish, archive, update, delete skill routes"
```

---

### Task 7: Run full verification

**Step 1: Format and tidy**

```bash
go fmt ./...
go mod tidy
go build -o server .
```

**Step 2: Run tests**

```bash
go test ./...
```

Expected: All tests pass.

**Step 3: Commit if any formatting changes**

---

## Status Flow Summary

```
draft ──publish──→ published ──archive──→ archived
  │                                          │
  └── edit/delete ──→ (any changes)          └── (terminal, no transitions out)
```

- **draft**: Can be edited, deleted, or published
- **published**: Immutable. Can only be archived. Subscriptions active.
- **archived**: Terminal state. Cannot be edited, published, or unarchived.
