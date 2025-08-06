package config

import (
	"log"
	"os"
)

type Config struct {
	Port     string
	DBPath   string
	UploadDir string
}

var AppConfig *Config

func InitConfig() {
	AppConfig = &Config{
		Port:     getEnv("PORT", "8080"),
		DBPath:   getEnv("DB_PATH", "./data.db"),
		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
	}

	// 创建上传目录
	if err := os.MkdirAll(AppConfig.UploadDir, 0755); err != nil {
		log.Printf("创建上传目录失败: %v", err)
	}

	log.Printf("配置初始化完成: %+v", AppConfig)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}