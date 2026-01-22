package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"generateTestData/backend/models"
	"io"
	"math"
	"net/http"
	"time"
)

type TaskService struct {
	dbService     *DatabaseService
	exportService *ExportService
}

func NewTaskService() *TaskService {
	return &TaskService{
		dbService:     NewDatabaseService(),
		exportService: NewExportService(),
	}
}

// 创建任务
func (s *TaskService) CreateTask(task *models.Task) error {
	// 验证任务配置
	if err := s.validateTask(task); err != nil {
		return err
	}

	// 保存到数据库
	return models.DB.Create(task).Error
}

// 获取任务列表
func (s *TaskService) GetTasks(page, pageSize int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// 获取总数
	models.DB.Model(&models.Task{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := models.DB.Preload("DataSource").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error

	return tasks, total, err
}

// 获取单个任务
func (s *TaskService) GetTask(id uint) (*models.Task, error) {
	var task models.Task
	err := models.DB.Preload("DataSource").First(&task, id).Error
	return &task, err
}

// 删除任务
func (s *TaskService) DeleteTask(id uint) error {
	return models.DB.Delete(&models.Task{}, id).Error
}

// 执行任务
func (s *TaskService) ExecuteTask(taskID uint) error {
	// 获取任务
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	// 检查任务状态
	if task.Status == models.TaskStatusRunning {
		return fmt.Errorf("任务正在执行中")
	}

	// 启动异步执行
	go s.executeTaskAsync(task)

	return nil
}

// 异步执行任务
func (s *TaskService) executeTaskAsync(task *models.Task) {
	// 更新任务状态为运行中
	s.updateTaskStatus(task.ID, models.TaskStatusRunning, 0, "")

	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("任务执行异常: %v", r)
			s.updateTaskStatus(task.ID, models.TaskStatusFailed, 0, errorMsg)
		}
	}()

	var err error
	switch task.Type {
	case models.TaskTypeDatabase:
		err = s.executeDatabaseTask(task)
	case models.TaskTypeJSON:
		err = s.executeJSONTask(task)
	case models.TaskTypeCSV:
		err = s.executeCSVTask(task)
	default:
		err = fmt.Errorf("不支持的任务类型: %s", task.Type)
	}

	if err != nil {
		s.updateTaskStatus(task.ID, models.TaskStatusFailed, 0, err.Error())
		return
	}

	// 更新完成状态
	now := time.Now()
	models.DB.Model(task).Updates(map[string]interface{}{
		"status":       models.TaskStatusCompleted,
		"progress":     100.0,
		"completed_at": &now,
		"error_msg":    "", // 清除错误信息
	})

	fmt.Printf("任务 %d 执行完成，耗时: %v\n", task.ID, time.Since(start))
}

// 执行数据库任务
func (s *TaskService) executeDatabaseTask(task *models.Task) error {
	// 获取数据源
	if task.DataSource == nil {
		return fmt.Errorf("数据源不能为空")
	}

	// 获取表结构
	tableInfo, err := s.dbService.GetTableStructure(task.DataSource, task.TableName)
	if err != nil {
		return fmt.Errorf("获取表结构失败: %v", err)
	}

	// 解析字段规则
	rules, err := task.GetFieldRules()
	if err != nil {
		return fmt.Errorf("解析字段规则失败: %v", err)
	}

	// 获取唯一字段
	uniqueFields, err := task.GetUniqueFields()
	if err != nil {
		return fmt.Errorf("解析唯一字段失败: %v", err)
	}

	// 为每个任务创建独立的生成器实例，避免并发冲突
	generatorService := NewGeneratorService(s.dbService)

	// 分批生成数据
	batchSize := int64(10000) // 每批1万条
	var generated int64

	for generated < task.Count {
		currentBatch := batchSize
		if generated+batchSize > task.Count {
			currentBatch = task.Count - generated
		}

		// 生成一批数据
		records := make([]map[string]interface{}, currentBatch)
		for i := int64(0); i < currentBatch; i++ {
			// 构造上下文
			context := map[string]interface{}{
				"rowIndex":   generated + i,
				"dataSource": task.DataSource,
			}

			record, err := generatorService.GenerateRecord(tableInfo, rules, uniqueFields, context)
			if err != nil {
				return fmt.Errorf("生成记录失败: %v", err)
			}
			records[i] = record
		}

		// 输出数据
		switch task.OutputType {
		case models.OutputTypeDatabase:
			err = s.exportService.InsertToDatabase(task.DataSource, task.TableName, records)
		case models.OutputTypeSQL:
			err = s.exportService.ExportToSQL(task.OutputPath, task.TableName, records, generated == 0)
		case models.OutputTypeMockServer:
			err = s.pushToMockServer(task, records)
		default:
			return fmt.Errorf("不支持的输出类型: %s", task.OutputType)
		}

		if err != nil {
			return fmt.Errorf("输出数据失败: %v", err)
		}

		generated += currentBatch
		progress := math.Round(float64(generated) / float64(task.Count) * 100)
		s.updateTaskProgress(task.ID, progress)
	}

	return nil
}

// 执行JSON任务
func (s *TaskService) executeJSONTask(task *models.Task) error {
	// 解析JSON结构
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(task.JSONSchema), &schema); err != nil {
		return fmt.Errorf("解析JSON结构失败: %v", err)
	}

	// 解析字段规则
	rules, err := task.GetFieldRules()
	if err != nil {
		return fmt.Errorf("解析字段规则失败: %v", err)
	}

	// 获取唯一字段
	uniqueFields, err := task.GetUniqueFields()
	if err != nil {
		return fmt.Errorf("解析唯一字段失败: %v", err)
	}

	// 为每个任务创建独立的生成器实例，避免并发冲突
	generatorService := NewGeneratorService(s.dbService)

	// 分批生成数据
	batchSize := int64(1000) // JSON数据每批1000条
	var generated int64

	for generated < task.Count {
		currentBatch := batchSize
		if generated+batchSize > task.Count {
			currentBatch = task.Count - generated
		}

		// 生成一批数据
		jsonObjects := make([]map[string]interface{}, currentBatch)
		for i := int64(0); i < currentBatch; i++ {
			// 构造上下文
			context := map[string]interface{}{
				"rowIndex":   generated + i,
				"dataSource": task.DataSource,
			}

			jsonObj, err := generatorService.GenerateJSON(schema, rules, uniqueFields, context)
			if err != nil {
				return fmt.Errorf("生成JSON对象失败: %v", err)
			}
			jsonObjects[i] = jsonObj
		}

		// 根据输出类型导出到文件
		switch task.OutputType {
		case models.OutputTypeJSON:
			err = s.exportService.ExportToJSON(task.OutputPath, jsonObjects, generated == 0)
			if err != nil {
				return fmt.Errorf("导出JSON失败: %v", err)
			}
		case models.OutputTypeTXT:
			err = s.exportService.ExportToTXT(task.OutputPath, jsonObjects, generated == 0)
			if err != nil {
				return fmt.Errorf("导出TXT失败: %v", err)
			}
		case models.OutputTypeMockServer:
			err = s.pushToMockServer(task, jsonObjects)
			if err != nil {
				return fmt.Errorf("推送至Mock Server失败: %v", err)
			}
		default:
			return fmt.Errorf("不支持的输出类型: %s", task.OutputType)
		}

		generated += currentBatch
		progress := math.Round(float64(generated) / float64(task.Count) * 100)
		s.updateTaskProgress(task.ID, progress)
	}

	return nil
}

// 推送数据到 Mock Server
func (s *TaskService) pushToMockServer(task *models.Task, data []map[string]interface{}) error {
	var config struct {
		URL   string `json:"url"`
		Token string `json:"token"`
		Type  string `json:"type"`
	}
	// 如果 Configuration 为空，尝试使用默认配置或报错
	if task.Configuration != "" {
		if err := json.Unmarshal([]byte(task.Configuration), &config); err != nil {
			return fmt.Errorf("解析Mock Server配置失败: %v", err)
		}
	}

	if config.URL == "" {
		// 尝试从 OutputPath 获取 URL (兼容性考虑)
		if task.OutputPath != "" && (task.OutputPath[:7] == "http://" || task.OutputPath[:8] == "https://") {
			config.URL = task.OutputPath
		} else {
			return fmt.Errorf("Mock Server URL不能为空")
		}
	}

	// 构造请求体
	payload := map[string]interface{}{
		"type": config.Type,
		"data": data,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.Token != "" {
		// 自动添加 Bearer 前缀（如果用户未提供）
		token := config.Token
		if len(token) < 7 || (token[:7] != "Bearer " && token[:7] != "bearer ") {
			token = "Bearer " + token
		}
		req.Header.Set("Authorization", token)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Mock Server返回错误 (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// 验证任务配置
func (s *TaskService) validateTask(task *models.Task) error {
	if task.Name == "" {
		return fmt.Errorf("任务名称不能为空")
	}

	if task.Count <= 0 {
		return fmt.Errorf("生成数量必须大于0")
	}

	switch task.Type {
	case models.TaskTypeDatabase:
		if task.DataSourceID == nil {
			return fmt.Errorf("数据库任务必须指定数据源")
		}
		if task.TableName == "" {
			return fmt.Errorf("数据库任务必须指定表名")
		}
	case models.TaskTypeJSON:
		if task.JSONSchema == "" {
			return fmt.Errorf("JSON任务必须指定JSON结构")
		}
		if task.OutputPath == "" && task.OutputType != models.OutputTypeMockServer {
			return fmt.Errorf("JSON任务必须指定输出路径")
		}
	case models.TaskTypeCSV:
		if task.JSONSchema == "" {
			return fmt.Errorf("CSV任务必须指定列结构")
		}
		if task.OutputPath == "" && task.OutputType != models.OutputTypeMockServer {
			return fmt.Errorf("CSV任务必须指定输出路径")
		}
	default:
		return fmt.Errorf("不支持的任务类型: %s", task.Type)
	}

	return nil
}

// 更新任务状态
func (s *TaskService) updateTaskStatus(taskID uint, status models.TaskStatus, progress float64, errorMsg string) {
	updates := map[string]interface{}{
		"status":    status,
		"progress":  progress,
		"error_msg": errorMsg, // 总是更新错误信息，空字符串表示清除错误
	}
	models.DB.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates)
}

// 更新任务进度
func (s *TaskService) updateTaskProgress(taskID uint, progress float64) {
	models.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("progress", progress)
}

// 获取任务状态
func (s *TaskService) GetTaskStatus(taskID uint) (*models.Task, error) {
	var task models.Task
	err := models.DB.Select("id, status, progress, error_msg").First(&task, taskID).Error
	return &task, err
}

// 更新任务
func (s *TaskService) UpdateTask(task *models.Task) error {
	// 验证任务配置
	if err := s.validateTask(task); err != nil {
		return err
	}

	// 更新数据库中的任务
	return models.DB.Save(task).Error
}

// 导出任务规则模板
func (s *TaskService) ExportTaskTemplate(taskID uint, name, description string) (*models.TaskTemplate, error) {
	// 获取任务信息
	task, err := s.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// 创建模板
	template := &models.TaskTemplate{
		Name:        name,
		Description: description,
		Type:        task.Type,
		JSONSchema:  task.JSONSchema,
		FieldRules:  task.FieldRules,
	}

	// 保存模板到数据库
	if err := models.DB.Create(template).Error; err != nil {
		return nil, fmt.Errorf("保存模板失败: %v", err)
	}

	return template, nil
}

// 导入任务规则模板
func (s *TaskService) ImportTaskTemplate(template *models.TaskTemplate) error {
	// 保存模板到数据库
	return models.DB.Create(template).Error
}

// 获取规则模板列表
func (s *TaskService) GetTaskTemplates() ([]models.TaskTemplate, error) {
	var templates []models.TaskTemplate
	err := models.DB.Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// 删除规则模板
func (s *TaskService) DeleteTaskTemplate(templateID uint) error {
	return models.DB.Delete(&models.TaskTemplate{}, templateID).Error
}

// 生成预览数据
func (s *TaskService) GeneratePreviewData(task *models.Task) (interface{}, error) {
	// 验证任务配置
	if err := s.validateTask(task); err != nil {
		return nil, err
	}

	// 解析字段规则
	var fieldRules map[string]models.FieldRule
	if task.FieldRules != "" {
		if err := json.Unmarshal([]byte(task.FieldRules), &fieldRules); err != nil {
			return nil, fmt.Errorf("解析字段规则失败: %v", err)
		}
	}

	if task.Type == "database" {
		// 数据库任务预览
		return s.generateDatabasePreview(task, fieldRules)
	} else if task.Type == "json" {
		// JSON任务预览
		return s.generateJSONPreview(task, fieldRules)
	} else if task.Type == "csv" {
		// CSV任务预览
		return s.generateCSVPreview(task, fieldRules)
	}

	return nil, fmt.Errorf("不支持的任务类型: %s", task.Type)
}

// 生成CSV预览数据
func (s *TaskService) generateCSVPreview(task *models.Task, fieldRules map[string]models.FieldRule) (interface{}, error) {
	// 解析列结构
	var columns []models.ColumnInfo
	if err := json.Unmarshal([]byte(task.JSONSchema), &columns); err != nil {
		return nil, fmt.Errorf("解析CSV列结构失败: %v", err)
	}

	// 构造 TableInfo
	tableInfo := &models.TableInfo{
		TableName: "csv_preview",
		Columns:   columns,
	}

	// 为预览创建独立的生成器实例
	generatorService := NewGeneratorService(s.dbService)

	// 生成一条数据
	context := map[string]interface{}{
		"rowIndex":   int64(0),
		"dataSource": task.DataSource,
	}
	data, err := generatorService.GenerateRecord(tableInfo, fieldRules, []string{}, context)
	if err != nil {
		return nil, fmt.Errorf("生成数据失败: %v", err)
	}

	return data, nil
}

// 生成数据库预览数据
func (s *TaskService) generateDatabasePreview(task *models.Task, fieldRules map[string]models.FieldRule) (interface{}, error) {
	// 获取数据源
	var dataSource models.DataSource
	if err := models.DB.First(&dataSource, task.DataSourceID).Error; err != nil {
		return nil, fmt.Errorf("获取数据源失败: %v", err)
	}

	// 获取表结构
	tableInfo, err := s.dbService.GetTableStructure(&dataSource, task.TableName)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %v", err)
	}

	// 为预览创建独立的生成器实例
	generatorService := NewGeneratorService(s.dbService)

	// 生成一条数据
	context := map[string]interface{}{
		"rowIndex":   int64(0),
		"dataSource": &dataSource,
	}
	data, err := generatorService.GenerateRecord(tableInfo, fieldRules, []string{}, context)
	if err != nil {
		return nil, fmt.Errorf("生成数据失败: %v", err)
	}

	return data, nil
}

// 生成JSON预览数据
func (s *TaskService) generateJSONPreview(task *models.Task, fieldRules map[string]models.FieldRule) (interface{}, error) {
	// 解析JSON结构
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(task.JSONSchema), &schema); err != nil {
		return nil, fmt.Errorf("解析JSON结构失败: %v", err)
	}

	// 为预览创建独立的生成器实例
	generatorService := NewGeneratorService(s.dbService)

	// 生成一条数据
	context := map[string]interface{}{
		"rowIndex":   int64(0),
		"dataSource": task.DataSource,
	}
	data, err := generatorService.GenerateJSON(schema, fieldRules, []string{}, context)
	if err != nil {
		return nil, fmt.Errorf("生成数据失败: %v", err)
	}

	return data, nil
}

// 执行CSV任务
func (s *TaskService) executeCSVTask(task *models.Task) error {
	// 解析列结构 (复用JSONSchema字段存储列信息)
	// 格式: [{"name": "col1", "type": "string"}, ...]
	var columns []models.ColumnInfo
	if err := json.Unmarshal([]byte(task.JSONSchema), &columns); err != nil {
		return fmt.Errorf("解析CSV列结构失败: %v", err)
	}

	// 提取列名作为表头
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Name
	}

	// 解析字段规则
	rules, err := task.GetFieldRules()
	if err != nil {
		return fmt.Errorf("解析字段规则失败: %v", err)
	}

	// 获取唯一字段
	uniqueFields, err := task.GetUniqueFields()
	if err != nil {
		return fmt.Errorf("解析唯一字段失败: %v", err)
	}

	// 为每个任务创建独立的生成器实例
	generatorService := NewGeneratorService(s.dbService)

	// 构造 TableInfo 供生成器使用
	tableInfo := &models.TableInfo{
		TableName: "csv_export",
		Columns:   columns,
	}

	// 分批生成数据
	batchSize := int64(5000) // CSV每批5000条
	var generated int64

	for generated < task.Count {
		currentBatch := batchSize
		if generated+batchSize > task.Count {
			currentBatch = task.Count - generated
		}

		// 生成一批数据
		records := make([]map[string]interface{}, currentBatch)
		for i := int64(0); i < currentBatch; i++ {
			// 构造上下文
			context := map[string]interface{}{
				"rowIndex":   generated + i,
				"dataSource": task.DataSource,
			}

			record, err := generatorService.GenerateRecord(tableInfo, rules, uniqueFields, context)
			if err != nil {
				return fmt.Errorf("生成记录失败: %v", err)
			}
			records[i] = record
		}

		// 导出数据
		switch task.OutputType {
		case models.OutputTypeCSV:
			err = s.exportService.ExportToCSV(task.OutputPath, headers, records, generated == 0)
			if err != nil {
				return fmt.Errorf("导出CSV失败: %v", err)
			}
		case models.OutputTypeMockServer:
			err = s.pushToMockServer(task, records)
			if err != nil {
				return fmt.Errorf("推送至Mock Server失败: %v", err)
			}
		default:
			// 默认为CSV (兼容旧数据)
			err = s.exportService.ExportToCSV(task.OutputPath, headers, records, generated == 0)
			if err != nil {
				return fmt.Errorf("导出CSV失败: %v", err)
			}
		}

		generated += currentBatch
		progress := math.Round(float64(generated) / float64(task.Count) * 100)
		s.updateTaskProgress(task.ID, progress)
	}

	return nil
}
