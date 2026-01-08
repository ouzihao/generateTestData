package main

import (
	"generateTestData/backend/config"
	"generateTestData/backend/controllers"
	"generateTestData/backend/models"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化数据库
	models.InitDB()

	// 创建Gin引擎
	r := gin.Default()

	// 设置CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件服务
	r.Static("/static", "./frontend/dist")
	r.StaticFile("/", "./frontend/dist/index.html")

	// 创建控制器实例
	dataSourceController := controllers.NewDataSourceController()
	taskController := controllers.NewTaskController()
	fileController := controllers.NewFileController()

	// API路由组
	api := r.Group("/api")
	{
		// 数据源管理
		datasource := api.Group("/datasource")
		{
			datasource.GET("", dataSourceController.List)
			datasource.POST("", dataSourceController.Create)
			datasource.GET("/:id", dataSourceController.Get)
			datasource.PUT("/:id", dataSourceController.Update)
			datasource.DELETE("/:id", dataSourceController.Delete)
			datasource.POST("/test", dataSourceController.TestConnection)
			datasource.GET("/tables/:id", dataSourceController.GetTables)
			datasource.GET("/table/:id/:table", dataSourceController.GetTableStructure)
		}

		// 任务管理
		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskController.Create)
			tasks.GET("", taskController.List)
			tasks.GET("/:id", taskController.Get)
			tasks.PUT("/:id", taskController.Update)
			tasks.POST("/:id/execute", taskController.Execute)
			tasks.GET("/:id/status", taskController.GetStatus)
			tasks.DELETE("/:id", taskController.Delete)
			tasks.POST("/preview", taskController.Preview)
			tasks.POST("/:id/export-template", taskController.ExportTemplate)
		}

		// 规则模板管理
		templates := api.Group("/templates")
		{
			templates.GET("", taskController.GetTemplates)
			templates.POST("/import", taskController.ImportTemplate)
			templates.DELETE("/:id", taskController.DeleteTemplate)
		}

		// 文件下载
		api.GET("/download/:filename", fileController.Download)
	}

	log.Printf("服务器启动在端口 :%s", config.AppConfig.Port)
	r.Run(":" + config.AppConfig.Port) // 使用配置中的端口
}
