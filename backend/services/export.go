package services

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"generateTestData/backend/config"
	"generateTestData/backend/models"
	"os"
	"path/filepath"
	"strings"
)

type ExportService struct{}

func NewExportService() *ExportService {
	return &ExportService{}
}

// ensureFileExtension 确保文件名有正确的后缀，如果没有则自动添加
// 参数：
//   - fileName: 用户输入的文件名
//   - extension: 期望的文件后缀（不含点号，如 "sql", "json", "txt"）
//
// 返回：带有正确后缀的文件名
func ensureFileExtension(fileName, extension string) string {
	if fileName == "" {
		return fileName
	}

	// 规范化扩展名（转为小写）
	extension = strings.ToLower(extension)
	fileNameLower := strings.ToLower(fileName)

	// 如果文件名已经有该后缀（不区分大小写），直接返回原文件名（保持原大小写）
	if strings.HasSuffix(fileNameLower, "."+extension) {
		return fileName
	}

	// 查找最后一个点号的位置
	idx := strings.LastIndex(fileName, ".")

	// 如果找到点号且不在开头或末尾，说明有其他后缀，替换为正确的后缀
	if idx > 0 && idx < len(fileName)-1 {
		return fileName[:idx] + "." + extension
	}

	// 如果没有后缀或点号在开头/末尾，添加后缀
	return fileName + "." + extension
}

// 插入数据到数据库
func (s *ExportService) InsertToDatabase(dataSource *models.DataSource, tableName string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	// 连接数据库
	db, err := s.connectDatabase(dataSource)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 构建插入SQL
	firstRecord := records[0]
	columns := make([]string, 0, len(firstRecord))
	for column := range firstRecord {
		columns = append(columns, column)
	}

	// 构建占位符
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 准备语句
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer stmt.Close()

	// 批量插入
	for _, record := range records {
		values := make([]interface{}, len(columns))
		for i, column := range columns {
			values[i] = record[column]
		}

		_, err = stmt.Exec(values...)
		if err != nil {
			return fmt.Errorf("插入数据失败: %v", err)
		}
	}

	return nil
}

// 导出为SQL文件（使用批量INSERT提高性能）
func (s *ExportService) ExportToSQL(fileName, tableName string, records []map[string]interface{}, isFirst bool) error {
	if len(records) == 0 {
		return nil
	}

	// 确保文件名有正确的后缀
	fileName = ensureFileExtension(fileName, "sql")

	// 自动拼接文件路径
	filePath := filepath.Join(config.AppConfig.GenerateDir, fileName)

	// 打开文件
	var file *os.File
	var err error
	if isFirst {
		file, err = os.Create(filePath)
	} else {
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	}
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 获取列名
	firstRecord := records[0]
	columns := make([]string, 0, len(firstRecord))
	for column := range firstRecord {
		columns = append(columns, column)
	}

	// 批量INSERT的每批大小（避免SQL语句过长）
	batchSize := 1000
	if len(records) < batchSize {
		batchSize = len(records)
	}

	// 分批写入批量INSERT语句
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		// 构建批量INSERT语句
		valuesList := make([]string, 0, end-i)
		for j := i; j < end; j++ {
			record := records[j]
			values := make([]string, len(columns))
			for k, column := range columns {
				value := record[column]
				if value == nil {
					values[k] = "NULL"
				} else {
					switch v := value.(type) {
					case string:
						values[k] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
					default:
						values[k] = fmt.Sprintf("%v", v)
					}
				}
			}
			valuesList = append(valuesList, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
		}

		// 写入批量INSERT语句
		sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;\n",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(valuesList, ", "))

		_, err = file.WriteString(sqlStr)
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	return nil
}

// 导出为TXT文件（每行一个JSON字符串）
func (s *ExportService) ExportToTXT(fileName string, jsonObjects []map[string]interface{}, isFirst bool) error {
	if len(jsonObjects) == 0 {
		return nil
	}

	// 确保文件名有正确的后缀
	fileName = ensureFileExtension(fileName, "txt")

	// 自动拼接文件路径
	filePath := filepath.Join(config.AppConfig.GenerateDir, fileName)

	// 打开文件
	var file *os.File
	var err error
	if isFirst {
		file, err = os.Create(filePath)
	} else {
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	}
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 写入JSON对象，每行一个
	for i, obj := range jsonObjects {
		jsonData, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("序列化JSON失败: %v", err)
		}

		// 写入JSON字符串和换行符
		if i == len(jsonObjects)-1 {
			_, err = file.WriteString(string(jsonData))
		} else {
			_, err = file.WriteString(string(jsonData) + "\n")
		}
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	return nil
}

// 导出为JSON文件
func (s *ExportService) ExportToJSON(fileName string, jsonObjects []map[string]interface{}, isFirst bool) error {
	if len(jsonObjects) == 0 {
		return nil
	}

	// 确保文件名有正确的后缀
	fileName = ensureFileExtension(fileName, "json")

	// 自动拼接文件路径
	filePath := filepath.Join(config.AppConfig.GenerateDir, fileName)

	// 打开文件
	var file *os.File
	var err error
	if isFirst {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
		// 写入数组开始符号
		_, err = file.WriteString("[\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	} else {
		file, err = os.OpenFile(filePath, os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("打开文件失败: %v", err)
		}
		// 移动到文件末尾前2个字符（去掉最后的\n]）
		stat, err := file.Stat()
		if err != nil {
			return fmt.Errorf("获取文件信息失败: %v", err)
		}
		_, err = file.Seek(stat.Size()-2, 0)
		if err != nil {
			return fmt.Errorf("移动文件指针失败: %v", err)
		}
		// 写入逗号
		_, err = file.WriteString(",\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}
	defer file.Close()

	// 写入JSON对象
	for i, obj := range jsonObjects {
		jsonData, err := json.MarshalIndent(obj, "  ", "  ")
		if err != nil {
			return fmt.Errorf("序列化JSON失败: %v", err)
		}

		// 添加缩进
		lines := strings.Split(string(jsonData), "\n")
		for j, line := range lines {
			if j == 0 {
				_, err = file.WriteString("  " + line + "\n")
			} else {
				_, err = file.WriteString("  " + line + "\n")
			}
			if err != nil {
				return fmt.Errorf("写入文件失败: %v", err)
			}
		}

		// 如果不是最后一个对象，添加逗号
		if i < len(jsonObjects)-1 {
			_, err = file.WriteString(",\n")
			if err != nil {
				return fmt.Errorf("写入文件失败: %v", err)
			}
		}
	}

	// 写入数组结束符号
	_, err = file.WriteString("\n]")
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// 连接数据库
func (s *ExportService) connectDatabase(dataSource *models.DataSource) (*sql.DB, error) {
	var dsn string
	switch dataSource.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dataSource.Username, dataSource.Password, dataSource.Host, dataSource.Port, dataSource.Database)
		return sql.Open("mysql", dsn)
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dataSource.Host, dataSource.Port, dataSource.Username, dataSource.Password, dataSource.Database)
		return sql.Open("postgres", dsn)
	case "sqlite":
		return sql.Open("sqlite3", dataSource.Database)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dataSource.Type)
	}
}

// ExportToCSV 导出为CSV文件
func (s *ExportService) ExportToCSV(fileName string, headers []string, records []map[string]interface{}, isFirstBatch bool) error {

	// 确保文件名有正确的后缀
	fileName = ensureFileExtension(fileName, "csv")

	// 自动拼接文件路径
	filePath := filepath.Join(config.AppConfig.GenerateDir, fileName)

	// 打开文件
	flag := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	if isFirstBatch {
		flag = os.O_TRUNC | os.O_WRONLY | os.O_CREATE
	}

	file, err := os.OpenFile(filePath, flag, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 写入BOM头，防止中文乱码 (仅首次写入)
	if isFirstBatch {
		file.WriteString("\xEF\xBB\xBF")
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头 (仅首次写入)
	if isFirstBatch {
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("写入表头失败: %v", err)
		}
	}

	// 写入数据
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			val := record[header]
			// 处理不同类型的值
			switch v := val.(type) {
			case nil:
				row[i] = ""
			case string:
				row[i] = v
			case float64:
				// 去除多余的小数点0
				str := fmt.Sprintf("%f", v)
				row[i] = strings.TrimRight(strings.TrimRight(str, "0"), ".")
			default:
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入数据失败: %v", err)
		}
	}

	return nil
}
