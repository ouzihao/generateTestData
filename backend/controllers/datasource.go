package controllers

import (
	"generateTestData/backend/models"
	"generateTestData/backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DataSourceController struct {
	dbService *services.DatabaseService
}

func NewDataSourceController() *DataSourceController {
	return &DataSourceController{
		dbService: services.NewDatabaseService(),
	}
}

// 创建数据源
func (c *DataSourceController) Create(ctx *gin.Context) {
	var dataSource models.DataSource
	if err := ctx.ShouldBindJSON(&dataSource); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试连接
	if err := c.dbService.TestConnection(&dataSource); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接失败: " + err.Error()})
		return
	}

	// 保存到数据库
	if err := models.DB.Create(&dataSource).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存数据源失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": dataSource})
}

// 获取数据源列表
func (c *DataSourceController) List(ctx *gin.Context) {
	var dataSources []models.DataSource
	if err := models.DB.Find(&dataSources).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取数据源列表失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": dataSources})
}

// 获取单个数据源
func (c *DataSourceController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var dataSource models.DataSource
	if err := models.DB.First(&dataSource, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "数据源不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": dataSource})
}

// 更新数据源
func (c *DataSourceController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var dataSource models.DataSource
	if err := models.DB.First(&dataSource, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "数据源不存在"})
		return
	}

	if err := ctx.ShouldBindJSON(&dataSource); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试连接
	if err := c.dbService.TestConnection(&dataSource); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接失败: " + err.Error()})
		return
	}

	// 更新数据库
	if err := models.DB.Save(&dataSource).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新数据源失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": dataSource})
}

// 删除数据源
func (c *DataSourceController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := models.DB.Delete(&models.DataSource{}, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除数据源失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 测试连接
func (c *DataSourceController) TestConnection(ctx *gin.Context) {
	var dataSource models.DataSource
	if err := ctx.ShouldBindJSON(&dataSource); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.dbService.TestConnection(&dataSource); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "连接失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "连接成功"})
}

// 获取表列表
func (c *DataSourceController) GetTables(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var dataSource models.DataSource
	if err := models.DB.First(&dataSource, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "数据源不存在"})
		return
	}

	tables, err := c.dbService.GetTables(&dataSource)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取表列表失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": tables})
}

// 获取表结构
func (c *DataSourceController) GetTableStructure(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	tableName := ctx.Param("table")
	if tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "表名不能为空"})
		return
	}

	var dataSource models.DataSource
	if err := models.DB.First(&dataSource, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "数据源不存在"})
		return
	}

	tableInfo, err := c.dbService.GetTableStructure(&dataSource, tableName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取表结构失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": tableInfo})
}