package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/service"
)

type SkillHandler struct {
	skillService *service.SkillService
}

func NewSkillHandler(skillService *service.SkillService) *SkillHandler {
	return &SkillHandler{skillService: skillService}
}

// CreateSkillRequest 创建技能的请求体
type CreateSkillRequest struct {
	Name          string                `json:"name" binding:"required"`
	Description   string                `json:"description"`
	Content       string                `json:"content" binding:"required"`
	Visibility    model.SkillVisibility `json:"visibility" example:"private"`
	Status        model.SkillStatus     `json:"status" example:"draft"`
	Category      string                `json:"category"`
	Tags          []string              `json:"tags"`
	IsFree        bool                  `json:"is_free" example:"true"`
	UsesSystemLLM bool                  `json:"uses_system_llm" example:"true"`
	MaxTokens     int                   `json:"max_tokens" example:"1000"`
	Parameters    model.SkillParameters `json:"parameters"`
	ScheduleCron  *string               `json:"schedule_cron"`
}

// CreateSkill godoc
// @Summary      Create skill
// @Description  Create a new AI skill
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      CreateSkillRequest  true  "Skill information"
// @Success      201  {object}  model.Skill  "Created skill"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills [post]
func (h *SkillHandler) CreateSkill(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建技能数据
	skillData := model.Skill{
		Name:          req.Name,
		Description:   req.Description,
		Content:       req.Content,
		Visibility:    req.Visibility,
		Status:        req.Status,
		Category:      req.Category,
		Tags:          req.Tags,
		IsFree:        req.IsFree,
		UsesSystemLLM: req.UsesSystemLLM,
		MaxTokens:     req.MaxTokens,
		Parameters:    req.Parameters,
		ScheduleCron:  req.ScheduleCron,
	}

	skill, err := h.skillService.CreateSkill(userID.(uint), skillData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, skill)
}

// GetSkill godoc
// @Summary      Get skill details
// @Description  Get skill details by ID
// @Tags         skills
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Skill ID"
// @Success      200  {object}  model.Skill  "Skill details"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      404  {object}  map[string]string  "技能不存在"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/{id} [get]
func (h *SkillHandler) GetSkill(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skill id"})
		return
	}

	// 尝试获取用户ID（如果已认证）
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		userID = userIDVal.(uint)
	}

	skill, err := h.skillService.GetSkillByIDWithVisibility(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}

	c.JSON(http.StatusOK, skill)
}

// GetMySkills godoc
// @Summary      Get my skills
// @Description  Get all skills created by the current user
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Skill  "Skill list"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/me [get]
func (h *SkillHandler) GetMySkills(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	skills, err := h.skillService.GetUserSkills(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skills)
}

// UpdateSkillRequest 更新技能的请求体
type UpdateSkillRequest struct {
	Name          *string                `json:"name"`
	Description   *string                `json:"description"`
	Content       *string                `json:"content"`
	Visibility    *model.SkillVisibility `json:"visibility"`
	Status        *model.SkillStatus     `json:"status"`
	Category      *string                `json:"category"`
	Tags          *[]string              `json:"tags"`
	IsFree        *bool                  `json:"is_free"`
	UsesSystemLLM *bool                  `json:"uses_system_llm"`
	MaxTokens     *int                   `json:"max_tokens"`
	Parameters    *model.SkillParameters `json:"parameters"`
	ScheduleCron  *string                `json:"schedule_cron"`
}

// UpdateSkill godoc
// @Summary      Update skill
// @Description  Update skill information
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Skill ID"
// @Param        request  body      UpdateSkillRequest  true  "Update fields"
// @Success      200  {object}  model.Skill  "Updated skill"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      404  {object}  map[string]string  "技能不存在"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/{id} [put]
func (h *SkillHandler) UpdateSkill(c *gin.Context) {
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

	var req UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建更新数据
	updates := model.Skill{}
	if req.Name != nil {
		updates.Name = *req.Name
	}
	if req.Description != nil {
		updates.Description = *req.Description
	}
	if req.Content != nil {
		updates.Content = *req.Content
	}
	if req.Visibility != nil {
		updates.Visibility = *req.Visibility
	}
	if req.Status != nil {
		updates.Status = *req.Status
	}
	if req.Category != nil {
		updates.Category = *req.Category
	}
	if req.Tags != nil {
		updates.Tags = *req.Tags
	}
	if req.IsFree != nil {
		updates.IsFree = *req.IsFree
	}
	if req.UsesSystemLLM != nil {
		updates.UsesSystemLLM = *req.UsesSystemLLM
	}
	if req.MaxTokens != nil {
		updates.MaxTokens = *req.MaxTokens
	}
	if req.Parameters != nil {
		updates.Parameters = *req.Parameters
	}
	updates.ScheduleCron = req.ScheduleCron

	skill, err := h.skillService.UpdateSkill(userID.(uint), uint(id), updates)
	if err != nil {
		if err.Error() == "unauthorized: only skill creator can update" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, skill)
}

// DeleteSkill godoc
// @Summary      Delete skill
// @Description  Delete skill
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Skill ID"
// @Success      204  "删除成功"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      404  {object}  map[string]string  "技能不存在"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/{id} [delete]
func (h *SkillHandler) DeleteSkill(c *gin.Context) {
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

	err = h.skillService.DeleteSkill(userID.(uint), uint(id))
	if err != nil {
		if err.Error() == "unauthorized: only skill creator can delete" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchSkillsRequest 搜索技能的请求参数
type SearchSkillsRequest struct {
	Query string `form:"q" example:"weather"`
	Page  int    `form:"page" example:"1"`
	Size  int    `form:"size" example:"20"`
}

// SearchSkills godoc
// @Summary      Search skills
// @Description  Search public skills
// @Tags         skills
// @Accept       json
// @Produce      json
// @Param        q     query     string  false  "Search keyword"
// @Param        page  query     int     false  "Page number"  default(1)
// @Param        size  query     int     false  "Page size"  default(20)
// @Success      200  {object}  map[string]interface{}  "搜索结果"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/search [get]
func (h *SkillHandler) SearchSkills(c *gin.Context) {
	var req SearchSkillsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	skills, total, err := h.skillService.SearchPublicSkills(req.Query, req.Page, req.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"skills": skills,
		"total":  total,
		"page":   req.Page,
		"size":   req.Size,
	})
}

// SubscribeToSkillRequest 订阅技能的请求体
type SubscribeToSkillRequest struct {
	Config model.SkillExecutionConfig `json:"config" binding:"required"`
}

// SubscribeToSkill godoc
// @Summary      Subscribe to skill
// @Description  Subscribe to a published skill
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Skill ID"
// @Param        request  body      SubscribeToSkillRequest  true  "Subscription configuration"
// @Success      201  {object}  model.SkillSubscription  "Subscription created"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      404  {object}  map[string]string  "技能不存在"
// @Failure      409  {object}  map[string]string  "已订阅"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/{id}/subscribe [post]
func (h *SkillHandler) SubscribeToSkill(c *gin.Context) {
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

	var req SubscribeToSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.skillService.SubscribeToSkill(userID.(uint), uint(id), req.Config)
	if err != nil {
		switch err.Error() {
		case "skill is not published":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case "private skill can only be subscribed by creator":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "already subscribed to this skill":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case "LLM configuration required for non-free skill":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetUserSubscriptions godoc
// @Summary      Get user's subscriptions
// @Description  Get all skill subscriptions of the current user
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.SkillSubscription  "Subscription list"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /skills/subscriptions [get]
func (h *SkillHandler) GetUserSubscriptions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subscriptions, err := h.skillService.GetUserSubscriptions(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// GetSubscription godoc
// @Summary      Get subscription details
// @Description  Get subscription details by ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Subscription ID"
// @Success      200  {object}  model.SkillSubscription  "Subscription details"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      404  {object}  map[string]string  "订阅不存在"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /subscriptions/{id} [get]
func (h *SkillHandler) GetSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	subscription, err := h.skillService.GetSubscriptionByID(userID.(uint), uint(id))
	if err != nil {
		if err.Error() == "unauthorized: only subscription owner can view" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		}
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscriptionRequest 更新订阅的请求体
type UpdateSubscriptionRequest struct {
	Config       *model.SkillExecutionConfig `json:"config"`
	Status       *string                     `json:"status"`
	ScheduleCron *string                     `json:"schedule_cron"`
	TimeZone     *string                     `json:"time_zone"`
	BotID        *uint                       `json:"bot_id"`
	ChannelID    *uint                       `json:"channel_id"`
	NextRunAt    *time.Time                  `json:"next_run_at"`
}

// UpdateSubscription godoc
// @Summary      Update subscription configuration
// @Description  Update subscription configuration information
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Subscription ID"
// @Param        request  body      UpdateSubscriptionRequest  true  "Update fields"
// @Success      200  {object}  model.SkillSubscription  "Updated subscription"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      404  {object}  map[string]string  "订阅不存在"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /subscriptions/{id} [put]
func (h *SkillHandler) UpdateSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为SkillSubscription更新
	updates := model.SkillSubscription{}
	if req.Config != nil {
		updates.Config = *req.Config
	}
	if req.Status != nil {
		updates.Status = *req.Status
	}
	updates.ScheduleCron = req.ScheduleCron
	if req.TimeZone != nil {
		updates.TimeZone = *req.TimeZone
	}
	updates.BotID = req.BotID
	updates.ChannelID = req.ChannelID
	updates.NextRunAt = req.NextRunAt

	subscription, err := h.skillService.UpdateSubscription(userID.(uint), uint(id), updates)
	if err != nil {
		if err.Error() == "unauthorized: only subscription owner can update" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		}
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription godoc
// @Summary      Delete subscription (unsubscribe)
// @Description  Delete subscription (unsubscribe from skill)
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Subscription ID"
// @Success      204  "删除成功"
// @Failure      400  {object}  map[string]string  "请求无效"
// @Failure      401  {object}  map[string]string  "未授权"
// @Failure      403  {object}  map[string]string  "无权限"
// @Failure      404  {object}  map[string]string  "订阅不存在"
// @Failure      500  {object}  map[string]string  "服务器错误"
// @Router       /subscriptions/{id} [delete]
func (h *SkillHandler) DeleteSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	err = h.skillService.UnsubscribeFromSkill(userID.(uint), uint(id))
	if err != nil {
		if err.Error() == "unauthorized: only subscription owner can unsubscribe" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
