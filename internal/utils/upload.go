package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	maxFileSize = 5 * 1024 * 1024 // 5MB
	perfilDir   = "static/uploads/perfil"
)

type UploadedFile struct {
	Filename     string
	OriginalName string
	Path         string
	Size         int64
	MimeType     string
}

// ValidateImage - Valida que el archivo sea una imagen válida
func ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	if header.Size > maxFileSize {
		return fmt.Errorf("la imagen no puede superar los 5MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !validExts[ext] {
		return fmt.Errorf("formato no válido. Use: JPG, JPEG, PNG, GIF o WEBP")
	}

	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error leyendo archivo: %v", err)
	}

	mimeType := http.DetectContentType(buffer)
	if !strings.HasPrefix(mimeType, "image/") {
		return fmt.Errorf("el archivo no es una imagen válida")
	}

	file.Seek(0, 0)
	return nil
}

// SaveProfileImage - Guarda imagen de perfil
func SaveProfileImage(file multipart.File, header *multipart.FileHeader) (*UploadedFile, error) {
	if err := ValidateImage(file, header); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(perfilDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio: %v", err)
	}

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("profile_%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(perfilDir, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo: %v", err)
	}
	defer dst.Close()

	file.Seek(0, 0)

	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("error guardando archivo: %v", err)
	}

	return &UploadedFile{
		Filename:     filename,
		OriginalName: header.Filename,
		Path:         "/static/uploads/perfil/" + filename,
		Size:         size,
		MimeType:     header.Header.Get("Content-Type"),
	}, nil
}

// DeleteFile - Elimina un archivo del sistema
func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	if strings.HasPrefix(filePath, "/static/") {
		filePath = strings.TrimPrefix(filePath, "/static/")
		filePath = "static/" + filePath
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error eliminando archivo: %v", err)
	}
	return nil
}
