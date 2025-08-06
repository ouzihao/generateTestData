package utils

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 生成MD5哈希
func MD5(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// 生成随机数字字符串
func RandomNumber(length int) string {
	const charset = "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// 生成随机整数
func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// 生成随机浮点数
func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// 生成随机布尔值
func RandomBool() bool {
	return rand.Intn(2) == 1
}

// 生成随机日期
func RandomDate(start, end time.Time) time.Time {
	delta := end.Unix() - start.Unix()
	sec := rand.Int63n(delta) + start.Unix()
	return time.Unix(sec, 0)
}

// 验证正则表达式
func ValidateRegex(pattern string) error {
	_, err := regexp.Compile(pattern)
	return err
}

// 生成符合正则表达式的字符串（简单实现）
func GenerateFromRegex(pattern string) (string, error) {
	// 这是一个简化的实现，只处理一些基本的正则表达式
	switch {
	case pattern == "\\d+":
		return RandomNumber(RandomInt(1, 10)), nil
	case pattern == "\\w+":
		return RandomString(RandomInt(3, 10)), nil
	case strings.Contains(pattern, "[0-9]"):
		return RandomNumber(RandomInt(1, 5)), nil
	case strings.Contains(pattern, "[a-zA-Z]"):
		return RandomString(RandomInt(3, 8)), nil
	default:
		// 对于复杂的正则表达式，返回一个默认值
		return RandomString(8), nil
	}
}

// 字符串转整数
func StringToInt(s string, defaultValue int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defaultValue
}

// 字符串转浮点数
func StringToFloat(s string, defaultValue float64) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return defaultValue
}

// 字符串转布尔值
func StringToBool(s string, defaultValue bool) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return defaultValue
}

// 检查字符串是否在切片中
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// 从切片中随机选择一个元素
func RandomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	return choices[rand.Intn(len(choices))]
}

// 格式化文件大小
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// 生成时间戳
func Timestamp() int64 {
	return time.Now().Unix()
}

// 格式化时间
func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.Format(layout)
}