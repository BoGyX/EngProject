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
	return &UploadHandler{
		uploadDir: resolveUploadDir(),
	}
}

func (h *UploadHandler) BaseDir() string {
	return h.uploadDir
}

func resolveUploadDir() string {
	candidates := []string{
		"./uploads",
		filepath.Join(os.TempDir(), "english-learning", "uploads"),
	}

	for _, candidate := range candidates {
		if err := ensureUploadDirectory(candidate); err == nil {
			log.Printf("using upload dir: %s", candidate)
			return candidate
		} else {
			log.Printf("upload dir %s is unavailable: %v", candidate, err)
		}
	}

	fallback := "./uploads"
	log.Printf("falling back to upload dir %s without preflight guarantee", fallback)
	return fallback
}

func ensureUploadDirectory(baseDir string) error {
	for _, subdir := range []string{"images", "audio"} {
		if err := os.MkdirAll(filepath.Join(baseDir, subdir), 0755); err != nil {
			return err
		}
	}

	testFilePath := filepath.Join(baseDir, ".write-test")
	if err := os.WriteFile(testFilePath, []byte("ok"), 0644); err != nil {
		return err
	}

	return os.Remove(testFilePath)
}

func (h *UploadHandler) ensureReady() error {
	return ensureUploadDirectory(h.uploadDir)
}

// UploadImage uploads an image.
func (h *UploadHandler) UploadImage(c *gin.Context) {
	if err := h.ensureReady(); err != nil {
		log.Printf("image upload directory is not ready (%s): %v", h.uploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload storage is unavailable"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("upload image form error: %v", err)
		log.Printf("content-type: %s", c.GetHeader("Content-Type"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size: 5MB"})
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	targetPath := filepath.Join(h.uploadDir, "images", filename)

	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		log.Printf("failed to save image upload to %s: %v", targetPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      fmt.Sprintf("/uploads/images/%s", filename),
		"filename": filename,
		"message":  "Image uploaded successfully",
	})
}

// UploadAudio uploads an audio file.
func (h *UploadHandler) UploadAudio(c *gin.Context) {
	if err := h.ensureReady(); err != nil {
		log.Printf("audio upload directory is not ready (%s): %v", h.uploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload storage is unavailable"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".m4a": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: mp3, wav, ogg, m4a"})
		return
	}

	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size: 10MB"})
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	targetPath := filepath.Join(h.uploadDir, "audio", filename)

	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		log.Printf("failed to save audio upload to %s: %v", targetPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      fmt.Sprintf("/uploads/audio/%s", filename),
		"filename": filename,
		"message":  "Audio uploaded successfully",
	})
}

// DeleteFile deletes an uploaded file.
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

	subdir := "audio"
	if fileType == "image" {
		subdir = "images"
	}

	targetPath := filepath.Join(h.uploadDir, subdir, filename)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if err := os.Remove(targetPath); err != nil {
		log.Printf("failed to delete upload %s: %v", targetPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
