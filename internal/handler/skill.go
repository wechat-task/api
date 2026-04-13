package handler

import (
	"net/http"
	"strconv"

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
// @Summary      创建新技能
// @Description  创建新的AI技能
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      CreateSkillRequest  true  "技能信息"
// @Success      201  {object}  model.Skill  "创建成功的技能"
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
// @Summary      获取技能详情
// @Description  根据ID获取技能详情
// @Tags         skills
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "技能ID"
// @Success      200  {object}  model.Skill  "技能详情"
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

	skill, err := h.skillService.GetSkillByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}

	c.JSON(http.StatusOK, skill)
}

// GetMySkills godoc
// @Summary      获取我的技能列表
// @Description  获取当前用户创建的所有技能
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Skill  "技能列表"
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
// @Summary      更新技能
// @Description  更新技能信息
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "技能ID"
// @Param        request  body      UpdateSkillRequest  true  "更新字段"
// @Success      200  {object}  model.Skill  "更新后的技能"
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
// @Summary      删除技能
// @Description  删除技能
// @Tags         skills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "技能ID"
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
// @Summary      搜索技能
// @Description  搜索公开技能
// @Tags         skills
// @Accept       json
// @Produce      json
// @Param        q     query     string  false  "搜索关键词"
// @Param        page  query     int     false  "页码"  default(1)
// @Param        size  query     int     false  "每页数量"  default(20)
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
