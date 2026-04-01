package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler() *UploadHandler {
	uploadDir := "./uploads"
	// Создаем директории для загрузок
	os.MkdirAll(filepath.Join(uploadDir, "images"), 0755)
	os.MkdirAll(filepath.Join(uploadDir, "audio"), 0755)
	
	return &UploadHandler{
		uploadDir: uploadDir,
	}
}

// UploadImage загружает изображение
// @Summary Загрузить изображение
// @Description Загружает изображение для карточки
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Файл изображения"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		// Логируем детали ошибки
		log.Printf("❌ Upload error: %v", err)
		log.Printf("📋 Content-Type: %s", c.GetHeader("Content-Type"))
		log.Printf("📋 Request Method: %s", c.Request.Method)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Проверяем расширение файла
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	// Проверяем размер файла (макс 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size: 5MB"})
		return
	}

	// Генерируем уникальное имя файла
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	filepath := filepath.Join(h.uploadDir, "images", filename)

	// Сохраняем файл
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Возвращаем URL файла
	fileURL := fmt.Sprintf("/uploads/images/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": filename,
		"message":  "Image uploaded successfully",
	})
}

// UploadAudio загружает аудио файл
// @Summary Загрузить аудио
// @Description Загружает аудио файл для карточки
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Аудио файл"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /upload/audio [post]
func (h *UploadHandler) UploadAudio(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Проверяем расширение файла
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".m4a": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: mp3, wav, ogg, m4a"})
		return
	}

	// Проверяем размер файла (макс 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size: 10MB"})
		return
	}

	// Генерируем уникальное имя файла
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	filepath := filepath.Join(h.uploadDir, "audio", filename)

	// Сохраняем файл
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Возвращаем URL файла
	fileURL := fmt.Sprintf("/uploads/audio/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": filename,
		"message":  "Audio uploaded successfully",
	})
}

// DeleteFile удаляет загруженный файл
// @Summary Удалить файл
// @Description Удаляет загруженный файл (изображение или аудио)
// @Tags upload
// @Produce json
// @Param type query string true "Тип файла (image или audio)"
// @Param filename query string true "Имя файла"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /upload/delete [delete]
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	fileType := c.Query("type")
	filename := c.Query("filename")

	if fileType == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing type or filename parameter"})
		return
	}

	if fileType != "image" && fileType != "audio" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type. Must be 'image' or 'audio'"})
		return
	}

	// Формируем путь к файлу
	var subdir string
	if fileType == "image" {
		subdir = "images"
	} else {
		subdir = "audio"
	}

	filepath := filepath.Join(h.uploadDir, subdir, filename)

	// Проверяем существование файла
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Удаляем файл
	if err := os.Remove(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
