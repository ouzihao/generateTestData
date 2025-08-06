package controllers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FileController struct{}

func NewFileController() *FileController {
	return &FileController{}
}

// 下载文件
func (c *FileController) Download(ctx *gin.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件名不能为空"})
		return
	}

	// 构建文件路径（假设文件在uploads目录下）
	filePath := filepath.Join("uploads", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 设置响应头
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/octet-stream")

	// 发送文件
	ctx.File(filePath)
}