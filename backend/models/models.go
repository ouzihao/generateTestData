package models

import (
	"encoding/json"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 任务状态枚举
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// 任务类型枚举
type TaskType string

const (
	TaskTypeDatabase TaskType = "database"
	TaskTypeJSON     TaskType = "json"
	TaskTypeCSV      TaskType = "csv"
)

// 输出类型枚举
type OutputType string

const (
	OutputTypeDatabase OutputType = "database"
	OutputTypeSQL      OutputType = "sql"
	OutputTypeJSON     OutputType = "json"
	OutputTypeTXT      OutputType = "txt"
	OutputTypeCSV      OutputType = "csv"
)

// 数据源配置
type DataSource struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"` // mysql, postgresql, sqlite
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Database  string    `json:"database"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 字段生成规则
type FieldRule struct {
	Type       string                 `json:"type"`       // fixed, sequence, random, range, regex, enum, reference, custom
	Value      interface{}            `json:"value"`      // 具体的值或配置
	Parameters map[string]interface{} `json:"parameters"` // 额外参数
}

// 任务配置
type Task struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	Name         string      `json:"name" gorm:"not null"`
	Type         TaskType    `json:"type" gorm:"not null"`
	DataSourceID *uint       `json:"dataSourceId"` // 数据库任务使用
	DataSource   *DataSource `json:"data_source" gorm:"foreignKey:DataSourceID"`
	TableName    string      `json:"tableName"`  // 数据库任务使用
	JSONSchema   string      `json:"jsonSchema"` // JSON任务使用，存储JSON结构定义
	FieldRules   string      `json:"fieldRules"` // 存储字段规则的JSON字符串
	Count        int64       `json:"count"`      // 生成数据数量
	OutputType   OutputType  `json:"outputType"`
	OutputPath   string      `json:"outputPath"`    // 输出文件名（不含路径，会自动保存到配置的生成目录）
	UniqueFields string      `json:"unique_fields"` // 不允许重复的字段，JSON数组格式
	Status       TaskStatus  `json:"status" gorm:"default:pending"`
	Progress     float64     `json:"progress" gorm:"default:0"`
	ErrorMsg     string      `json:"error_msg"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	CompletedAt  *time.Time  `json:"completed_at"`
}

// 表结构信息
type TableInfo struct {
	TableName string       `json:"table_name"`
	Columns   []ColumnInfo `json:"columns"`
}

// 列信息
type ColumnInfo struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Nullable        bool   `json:"nullable"`
	DefaultValue    string `json:"default_value"`
	IsPrimaryKey    bool   `json:"is_primary_key"`
	IsAutoIncrement bool   `json:"is_auto_increment"`
	MaxLength       int    `json:"max_length"`
}

// 任务执行结果
type TaskResult struct {
	TaskID         uint          `json:"task_id"`
	GeneratedCount int64         `json:"generated_count"`
	FilePath       string        `json:"file_path"`
	Duration       time.Duration `json:"duration"`
	CreatedAt      time.Time     `json:"created_at"`
}

// 解析字段规则
func (t *Task) GetFieldRules() (map[string]FieldRule, error) {
	var rules map[string]FieldRule
	if t.FieldRules == "" {
		return make(map[string]FieldRule), nil
	}
	err := json.Unmarshal([]byte(t.FieldRules), &rules)
	return rules, err
}

// 设置字段规则
func (t *Task) SetFieldRules(rules map[string]FieldRule) error {
	data, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	t.FieldRules = string(data)
	return nil
}

// 获取唯一字段列表
func (t *Task) GetUniqueFields() ([]string, error) {
	var fields []string
	if t.UniqueFields == "" {
		return fields, nil
	}
	err := json.Unmarshal([]byte(t.UniqueFields), &fields)
	return fields, err
}

// 设置唯一字段列表
func (t *Task) SetUniqueFields(fields []string) error {
	data, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	t.UniqueFields = string(data)
	return nil
}

// 任务规则模板
type TaskTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Type        TaskType  `json:"type" gorm:"not null"`
	JSONSchema  string    `json:"jsonSchema"` // JSON任务使用，存储JSON结构定义
	FieldRules  string    `json:"fieldRules"` // 存储字段规则的JSON字符串
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 解析字段规则
func (tt *TaskTemplate) GetFieldRules() (map[string]FieldRule, error) {
	var rules map[string]FieldRule
	if tt.FieldRules == "" {
		return make(map[string]FieldRule), nil
	}
	err := json.Unmarshal([]byte(tt.FieldRules), &rules)
	return rules, err
}

// 设置字段规则
func (tt *TaskTemplate) SetFieldRules(rules map[string]FieldRule) error {
	data, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	tt.FieldRules = string(data)
	return nil
}

// 初始化数据库
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}

	// 自动迁移
	err = DB.AutoMigrate(&DataSource{}, &Task{}, &TaskTemplate{})
	if err != nil {
		panic("数据库迁移失败: " + err.Error())
	}
}
