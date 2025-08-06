package controllers

import (
	"generateTestData/backend/models"
	"generateTestData/backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService *services.TaskService
}

func NewTaskController() *TaskController {
	return &TaskController{
		taskService: services.NewTaskService(),
	}
}

// 创建任务
func (c *TaskController) Create(ctx *gin.Context) {
	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.taskService.CreateTask(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": task})
}

// 获取任务列表
func (c *TaskController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	tasks, total, err := c.taskService.GetTasks(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务列表失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list":     tasks,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// 获取单个任务
func (c *TaskController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	task, err := c.taskService.GetTask(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": task})
}

// 删除任务
func (c *TaskController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := c.taskService.DeleteTask(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 执行任务
func (c *TaskController) Execute(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := c.taskService.ExecuteTask(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务已开始执行"})
}

// 获取任务状态
func (c *TaskController) GetStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	task, err := c.taskService.GetTaskStatus(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": task})
}

// 更新任务
func (c *TaskController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = uint(id)
	if err := c.taskService.UpdateTask(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": task})
}

// 导出任务规则模板
func (c *TaskController) ExportTemplate(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 接收模板信息
	var templateInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&templateInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := c.taskService.ExportTaskTemplate(uint(id), templateInfo.Name, templateInfo.Description)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": template, "message": "模板导出成功"})
}

// 导入任务规则模板
func (c *TaskController) ImportTemplate(ctx *gin.Context) {
	var template models.TaskTemplate
	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.taskService.ImportTaskTemplate(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "模板导入成功"})
}

// 获取规则模板列表
func (c *TaskController) GetTemplates(ctx *gin.Context) {
	templates, err := c.taskService.GetTaskTemplates()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取模板列表失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": templates})
}

// 删除规则模板
func (c *TaskController) DeleteTemplate(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := c.taskService.DeleteTaskTemplate(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除模板失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 生成预览数据
func (c *TaskController) Preview(ctx *gin.Context) {
	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 强制设置count为1，预览只生成一条数据
	task.Count = 1

	data, err := c.taskService.GeneratePreviewData(&task)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": data})
}