package services

import (
	"fmt"
	"generateTestData/backend/models"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	regen "github.com/zach-klippenstein/goregen"
)

type GeneratorService struct {
	uniqueValues     map[string]map[interface{}]bool // 用于存储唯一值
	sequenceCounters map[string]*big.Int             // 序列计数器，支持大整数
}

func NewGeneratorService() *GeneratorService {
	return &GeneratorService{
		uniqueValues:     make(map[string]map[interface{}]bool),
		sequenceCounters: make(map[string]*big.Int),
	}
}

// 生成单条数据库记录
func (g *GeneratorService) GenerateRecord(tableInfo *models.TableInfo, rules map[string]models.FieldRule, uniqueFields []string) (map[string]interface{}, error) {
	record := make(map[string]interface{})

	for _, column := range tableInfo.Columns {
		rule, exists := rules[column.Name]
		if !exists {
			// 如果没有规则，使用默认规则
			rule = g.getDefaultRule(column)
		}

		value, err := g.generateValue(column.Name, column.Type, rule, uniqueFields)
		if err != nil {
			return nil, fmt.Errorf("生成字段 %s 的值失败: %v", column.Name, err)
		}

		record[column.Name] = value
	}

	return record, nil
}

// 生成JSON对象
func (g *GeneratorService) GenerateJSON(schema map[string]interface{}, rules map[string]models.FieldRule, uniqueFields []string) (map[string]interface{}, error) {
	result, err := g.generateJSONValue("", schema, rules, uniqueFields)
	if err != nil {
		return nil, err
	}
	if jsonObj, ok := result.(map[string]interface{}); ok {
		return jsonObj, nil
	}
	return nil, fmt.Errorf("生成的结果不是有效的JSON对象")
}

// 生成值
func (g *GeneratorService) generateValue(fieldName, fieldType string, rule models.FieldRule, uniqueFields []string) (interface{}, error) {
	var value interface{}
	var err error

	// 为随机生成添加字段名信息
	if rule.Type == "random" || rule.Type == "" {
		if rule.Parameters == nil {
			rule.Parameters = make(map[string]interface{})
		}
		rule.Parameters["fieldName"] = fieldName
	}

	switch rule.Type {
	case "fixed":
		if val, ok := rule.Parameters["value"]; ok {
			value = val
		} else {
			value = rule.Value // 兼容旧格式
		}
	case "sequence", "increment":
		value, err = g.generateSequence(fieldName, rule)
	case "date_sequence":
		value, err = g.generateDateSequence(fieldName, rule)
	case "random":
		value, err = g.generateRandom(fieldType, rule)
	case "range":
		value, err = g.generateRange(fieldType, rule)
	case "regex":
		value, err = g.generateRegex(rule)
	case "enum":
		value, err = g.generateEnum(rule)
	case "uuid":
		value = g.generateUUID()
	case "custom":
		value, err = g.generateCustom(rule)
	default:
		// 为默认情况也添加字段名信息
		if rule.Parameters == nil {
			rule.Parameters = make(map[string]interface{})
		}
		rule.Parameters["fieldName"] = fieldName
		value, err = g.generateRandom(fieldType, rule)
	}

	if err != nil {
		return nil, err
	}

	// 检查唯一性约束
	if g.isUniqueField(fieldName, uniqueFields) {
		if g.isValueExists(fieldName, value) {
			// 如果值已存在，重新生成
			return g.generateValue(fieldName, fieldType, rule, uniqueFields)
		}
		g.addUniqueValue(fieldName, value)
	}

	return value, nil
}

// 生成序列值
func (g *GeneratorService) generateSequence(fieldName string, rule models.FieldRule) (interface{}, error) {
	// 解析起始值，支持字符串和数字
	var start *big.Int
	if startParam, ok := rule.Parameters["start"]; ok {
		switch v := startParam.(type) {
		case string:
			var success bool
			start, success = new(big.Int).SetString(v, 10)
			if !success {
				return nil, fmt.Errorf("invalid start value: %s", v)
			}
		case float64:
			start = big.NewInt(int64(v))
		case int:
			start = big.NewInt(int64(v))
		case int64:
			start = big.NewInt(v)
		default:
			start = big.NewInt(1)
		}
	} else {
		start = big.NewInt(1)
	}

	// 解析步长，支持字符串和数字
	var step *big.Int
	if stepParam, ok := rule.Parameters["step"]; ok {
		switch v := stepParam.(type) {
		case string:
			var success bool
			step, success = new(big.Int).SetString(v, 10)
			if !success {
				return nil, fmt.Errorf("invalid step value: %s", v)
			}
		case float64:
			step = big.NewInt(int64(v))
		case int:
			step = big.NewInt(int64(v))
		case int64:
			step = big.NewInt(v)
		default:
			step = big.NewInt(1)
		}
	} else {
		step = big.NewInt(1)
	}

	// 初始化或递增计数器
	if _, exists := g.sequenceCounters[fieldName]; !exists {
		g.sequenceCounters[fieldName] = new(big.Int).Set(start)
	} else {
		g.sequenceCounters[fieldName].Add(g.sequenceCounters[fieldName], step)
	}

	// 返回字符串格式的结果，保持大整数精度
	return g.sequenceCounters[fieldName].String(), nil
}

// 生成随机值
func (g *GeneratorService) generateRandom(fieldType string, rule models.FieldRule) (interface{}, error) {
	// 获取字段名用于日期识别
	fieldName := ""
	if name, ok := rule.Parameters["fieldName"].(string); ok {
		// 提取字段名的最后部分（去掉路径前缀）
		parts := strings.Split(name, ".")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			// 去掉数组标记
			lastPart = strings.Replace(lastPart, "[]", "", -1)
			fieldName = strings.ToLower(lastPart)
		}

	}

	switch strings.ToLower(fieldType) {
	case "int", "integer", "bigint", "smallint", "tinyint":
		return rand.Intn(1000000), nil
	case "varchar", "text", "char", "string":
		// 检查字段名是否包含日期相关关键词
		if strings.Contains(fieldName, "date") || strings.Contains(fieldName, "time") ||
			strings.Contains(fieldName, "created") || strings.Contains(fieldName, "updated") ||
			strings.Contains(fieldName, "birth") || strings.Contains(fieldName, "expire") {
			// 生成日期字符串
			result := g.generateRandomDate()
			if format, ok := rule.Parameters["format"].(string); ok {
				return result.Format(format), nil
			}
			return result.Format("2006-01-02"), nil
		}

		length := 10
		if len, ok := rule.Parameters["length"].(float64); ok {
			length = int(len)
		}
		return g.generateRandomString(length), nil
	case "decimal", "float", "double", "numeric":
		return rand.Float64() * 1000, nil
	case "date":
		// 如果有日期范围参数，使用日期范围生成
		if _, hasStart := rule.Parameters["start"]; hasStart {
			return g.generateDateRange(rule)
		}
		if _, hasEnd := rule.Parameters["end"]; hasEnd {
			return g.generateDateRange(rule)
		}
		// 否则使用默认随机日期
		result := g.generateRandomDate()
		if format, ok := rule.Parameters["format"].(string); ok {
			return result.Format(format), nil
		}
		return result.Format("2006-01-02"), nil
	case "datetime", "timestamp":
		// 如果有日期范围参数，使用日期范围生成
		if _, hasStart := rule.Parameters["start"]; hasStart {
			return g.generateDateRange(rule)
		}
		if _, hasEnd := rule.Parameters["end"]; hasEnd {
			return g.generateDateRange(rule)
		}
		// 否则使用默认随机日期时间
		result := g.generateRandomDate()
		if format, ok := rule.Parameters["format"].(string); ok {
			return result.Format(format), nil
		}
		return result.Format("2006-01-02 15:04:05"), nil
	case "boolean", "bool":
		return rand.Intn(2) == 1, nil
	default:
		return g.generateRandomString(10), nil
	}
}

// 生成范围值
func (g *GeneratorService) generateRange(fieldType string, rule models.FieldRule) (interface{}, error) {
	min, ok1 := rule.Parameters["min"]
	max, ok2 := rule.Parameters["max"]
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("范围规则需要min和max参数")
	}

	// 安全的类型转换函数
	convertToFloat := func(val interface{}) (float64, error) {
		switch v := val.(type) {
		case float64:
			return v, nil
		case string:
			return strconv.ParseFloat(v, 64)
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		default:
			return 0, fmt.Errorf("无法转换类型 %T 为 float64", val)
		}
	}

	minFloat, err := convertToFloat(min)
	if err != nil {
		return nil, fmt.Errorf("min参数转换失败: %v", err)
	}
	maxFloat, err := convertToFloat(max)
	if err != nil {
		return nil, fmt.Errorf("max参数转换失败: %v", err)
	}

	switch strings.ToLower(fieldType) {
	case "int", "integer", "bigint", "smallint", "tinyint":
		minVal := int(minFloat)
		maxVal := int(maxFloat)
		return rand.Intn(maxVal-minVal+1) + minVal, nil
	case "decimal", "float", "double", "numeric":
		return rand.Float64()*(maxFloat-minFloat) + minFloat, nil
	default:
		return nil, fmt.Errorf("字段类型 %s 不支持范围生成", fieldType)
	}
}

// 生成正则表达式值
func (g *GeneratorService) generateRegex(rule models.FieldRule) (interface{}, error) {
	pattern, ok := rule.Parameters["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("正则表达式规则需要pattern参数")
	}

	// 优先处理常见模式，避免goregen生成过长字符串
	if pattern == "\\d{11}" || pattern == "[0-9]{11}" {
		return g.generatePhone(), nil
	}
	if pattern == "1[3-9]\\d{9}" || pattern == "1[3-9][0-9]{9}" {
		return g.generatePhone(), nil
	}
	if pattern == "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}" {
		return g.generateEmail(), nil
	}

	// 转换不支持的转义序列
	// goregen 不支持 \d, \w, \s 等转义序列，需要转换为字符类
	convertedPattern := strings.ReplaceAll(pattern, "\\d", "[0-9]")
	convertedPattern = strings.ReplaceAll(convertedPattern, "\\w", "[a-zA-Z0-9_]")
	convertedPattern = strings.ReplaceAll(convertedPattern, "\\s", "[ \\t\\n\\r]")

	// 使用 goregen 生成符合正则的随机字符串
	result, err := regen.Generate(convertedPattern)
	if err != nil {
		// 最后回退到随机字符串
		return g.generateRandomString(10), fmt.Errorf("无法解析正则表达式 %s: %v，使用默认随机字符串", pattern, err)
	}

	return result, nil
}

// 生成枚举值
func (g *GeneratorService) generateEnum(rule models.FieldRule) (interface{}, error) {
	valuesStr, ok := rule.Parameters["values"].(string)
	if !ok {
		return nil, fmt.Errorf("枚举规则需要values参数")
	}

	// 将逗号分隔的字符串转换为数组
	values := strings.Split(strings.TrimSpace(valuesStr), ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	if len(values) == 0 || (len(values) == 1 && values[0] == "") {
		return nil, fmt.Errorf("枚举值不能为空")
	}

	return values[rand.Intn(len(values))], nil
}

// 生成自定义值
func (g *GeneratorService) generateCustom(rule models.FieldRule) (interface{}, error) {
	script, ok := rule.Parameters["script"].(string)
	if !ok {
		return nil, fmt.Errorf("自定义规则需要script参数")
	}

	// 简化实现，暂时返回随机字符串
	// 后续可以集成JavaScript引擎执行script
	_ = script // 避免未使用变量警告
	return g.generateRandomString(10), nil
}

// 生成JSON值（递归处理嵌套结构）
func (g *GeneratorService) generateJSONValue(path string, schema interface{}, rules map[string]models.FieldRule, uniqueFields []string) (interface{}, error) {
	switch v := schema.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			fieldPath := path
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += key

			generatedValue, err := g.generateJSONValue(fieldPath, value, rules, uniqueFields)
			if err != nil {
				return nil, err
			}
			result[key] = generatedValue
		}
		return result, nil

	case []interface{}:
		if len(v) == 0 {
			return []interface{}{}, nil
		}

		// 获取数组长度配置
		arrayLength := 3 // 默认长度
		if rule, exists := rules[path]; exists {
			if length, ok := rule.Parameters["length"].(float64); ok {
				arrayLength = int(length)
			}
		}

		result := make([]interface{}, arrayLength)
		// 使用统一的数组元素路径格式，与前端保持一致
		arrayElementPath := path + "[]"
		for i := 0; i < arrayLength; i++ {
			generatedValue, err := g.generateJSONValue(arrayElementPath, v[0], rules, uniqueFields)
			if err != nil {
				return nil, err
			}
			result[i] = generatedValue
		}
		return result, nil

	case string:
		// 基本类型，根据规则生成值
		// 对于JSON schema中的字符串值，统一按string类型处理
		if rule, exists := rules[path]; exists {
			return g.generateValue(path, "string", rule, uniqueFields)
		}
		return g.generateValue(path, "string", models.FieldRule{Type: "random"}, uniqueFields)

	case float64:
		// 处理数值类型（浮点数）
		if rule, exists := rules[path]; exists {
			return g.generateValue(path, "decimal", rule, uniqueFields)
		}
		return g.generateValue(path, "decimal", models.FieldRule{Type: "random"}, uniqueFields)

	case int:
		// 处理数值类型（整数）
		if rule, exists := rules[path]; exists {
			return g.generateValue(path, "int", rule, uniqueFields)
		}
		return g.generateValue(path, "int", models.FieldRule{Type: "random"}, uniqueFields)

	default:
		return v, nil
	}
}

// 获取默认规则
func (g *GeneratorService) getDefaultRule(column models.ColumnInfo) models.FieldRule {
	if column.IsAutoIncrement {
		return models.FieldRule{
			Type: "sequence",
			Parameters: map[string]interface{}{
				"start": 1,
				"step":  1,
			},
		}
	}

	return models.FieldRule{Type: "random"}
}

// 检查是否为唯一字段
func (g *GeneratorService) isUniqueField(fieldName string, uniqueFields []string) bool {
	for _, field := range uniqueFields {
		if field == fieldName {
			return true
		}
	}
	return false
}

// 检查值是否已存在
func (g *GeneratorService) isValueExists(fieldName string, value interface{}) bool {
	if values, exists := g.uniqueValues[fieldName]; exists {
		// uniqueValues  =>  map[string]map[interface{}]bool
		return values[value]
	}
	return false
}

// 添加唯一值
func (g *GeneratorService) addUniqueValue(fieldName string, value interface{}) {
	if _, exists := g.uniqueValues[fieldName]; !exists {
		g.uniqueValues[fieldName] = make(map[interface{}]bool)
	}
	g.uniqueValues[fieldName][value] = true
}

// 重置生成器状态
func (g *GeneratorService) Reset() {
	g.uniqueValues = make(map[string]map[interface{}]bool)
	// 注意：不重置序列计数器，保持序列的连续性
	// g.sequenceCounters = make(map[string]*big.Int)
}

// 重置序列计数器（仅在需要时调用）
func (g *GeneratorService) ResetSequenceCounters() {
	g.sequenceCounters = make(map[string]*big.Int)
}

// 生成UUID
func (g *GeneratorService) generateUUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rand.Uint32(),
		rand.Uint32()&0xffff,
		rand.Uint32()&0xffff,
		rand.Uint32()&0xffff,
		rand.Uint64()&0xffffffffffff)
}

// 生成随机字符串
func (g *GeneratorService) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// 生成随机日期
func (g *GeneratorService) generateRandomDate() time.Time {
	min := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Now().Unix()
	delta := max - min
	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

// 生成日期范围内的随机日期
func (g *GeneratorService) generateDateRange(rule models.FieldRule) (interface{}, error) {
	var startTime, endTime time.Time
	var err error

	// 解析开始时间
	if startStr, ok := rule.Parameters["start"].(string); ok {
		startTime, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			// 尝试解析带时间的格式
			startTime, err = time.Parse("2006-01-02 15:04:05", startStr)
			if err != nil {
				return nil, fmt.Errorf("无效的开始日期格式: %s", startStr)
			}
		}
	} else {
		startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// 解析结束时间
	if endStr, ok := rule.Parameters["end"].(string); ok {
		endTime, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			// 尝试解析带时间的格式
			endTime, err = time.Parse("2006-01-02 15:04:05", endStr)
			if err != nil {
				return nil, fmt.Errorf("无效的结束日期格式: %s", endStr)
			}
		}
	} else {
		endTime = time.Now()
	}

	// 确保开始时间小于结束时间
	if startTime.After(endTime) {
		return nil, fmt.Errorf("开始时间不能晚于结束时间")
	}

	// 生成随机时间
	delta := endTime.Unix() - startTime.Unix()
	if delta <= 0 {
		return startTime, nil
	}

	sec := rand.Int63n(delta) + startTime.Unix()
	resultTime := time.Unix(sec, 0)

	// 根据格式参数返回相应格式
	if format, ok := rule.Parameters["format"].(string); ok {
		return resultTime.Format(format), nil
	}

	// 默认返回 YYYY-MM-DD 格式
	return resultTime.Format("2006-01-02"), nil
}

// 生成日期序列
func (g *GeneratorService) generateDateSequence(fieldName string, rule models.FieldRule) (interface{}, error) {
	// 获取开始日期
	startStr, ok := rule.Parameters["start"].(string)
	if !ok {
		startStr = "2024-01-01"
	}

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, fmt.Errorf("无效的开始日期格式: %s", startStr)
	}

	// 获取步长（天数）
	stepDays := 1
	if step, ok := rule.Parameters["step"]; ok {
		switch v := step.(type) {
		case string:
			if stepInt, err := strconv.Atoi(v); err == nil {
				stepDays = stepInt
			}
		case float64:
			stepDays = int(v)
		case int:
			stepDays = v
		}
	}

	// 获取或初始化计数器
	counterKey := fmt.Sprintf("date_%s", fieldName)
	if _, exists := g.sequenceCounters[counterKey]; !exists {
		g.sequenceCounters[counterKey] = big.NewInt(0)
	}

	// 计算当前日期
	currentCount := g.sequenceCounters[counterKey].Int64()
	currentTime := startTime.AddDate(0, 0, int(currentCount*int64(stepDays)))

	// 递增计数器
	g.sequenceCounters[counterKey].Add(g.sequenceCounters[counterKey], big.NewInt(1))

	// 根据格式参数返回相应格式
	if format, ok := rule.Parameters["format"].(string); ok {
		return currentTime.Format(format), nil
	}

	// 默认返回 YYYY-MM-DD 格式
	return currentTime.Format("2006-01-02"), nil
}

// 生成随机手机号
func (g *GeneratorService) generatePhone() string {
	return fmt.Sprintf("1%d%08d", rand.Intn(9)+1, rand.Intn(100000000))
}

// 生成随机邮箱
func (g *GeneratorService) generateEmail() string {
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "qq.com"}
	username := g.generateRandomString(8)
	domain := domains[rand.Intn(len(domains))]
	return fmt.Sprintf("%s@%s", username, domain)
}
