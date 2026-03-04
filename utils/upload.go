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
	carruselDir = "static/uploads/carrusel"
	perfilDir   = "static/uploads/perfil"
)

type UploadedFile struct {
	Filename     string
	OriginalName string
	Path         string
	Size         int64
	MimeType     string
}

// ValidateImage - Validación genérica para imágenes
func ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	// Validar tamaño
	if header.Size > maxFileSize {
		return fmt.Errorf("la imagen no puede superar los 5MB")
	}

	// Validar extensión
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !validExts[ext] {
		return fmt.Errorf("formato no válido. Use: JPG, PNG o GIF")
	}

	// Validar tipo MIME
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error leyendo archivo: %v", err)
	}

	mimeType := http.DetectContentType(buffer)
	if !strings.HasPrefix(mimeType, "image/") {
		return fmt.Errorf("el archivo no es una imagen válida")
	}

	// Reposicionar el puntero al inicio para futuras lecturas
	file.Seek(0, 0)

	return nil
}

// ============================================
// FUNCIONES PARA CARRUSEL
// ============================================

// SaveUploadedFile - Guarda imagen para carrusel
func SaveUploadedFile(file multipart.File, header *multipart.FileHeader) (*UploadedFile, error) {
	// Validar archivo
	if err := ValidateImage(file, header); err != nil {
		return nil, err
	}

	// Crear directorio si no existe
	if err := os.MkdirAll(carruselDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio: %v", err)
	}

	// Generar nombre único
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("carrusel_%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(carruselDir, filename)

	// Guardar archivo
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo: %v", err)
	}
	defer dst.Close()

	// Posicionar el puntero al inicio (por si acaso)
	file.Seek(0, 0)

	// Copiar contenido
	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("error guardando archivo: %v", err)
	}

	return &UploadedFile{
		Filename:     filename,
		OriginalName: header.Filename,
		Path:         "/static/uploads/carrusel/" + filename,
		Size:         size,
		MimeType:     header.Header.Get("Content-Type"),
	}, nil
}

// ============================================
// FUNCIONES PARA PERFIL DE USUARIO
// ============================================

// SaveProfileImage - Guarda imagen de perfil con nombre específico
func SaveProfileImage(file multipart.File, header *multipart.FileHeader) (*UploadedFile, error) {
	// Validar archivo
	if err := ValidateImage(file, header); err != nil {
		return nil, err
	}

	// Crear directorio si no existe
	if err := os.MkdirAll(perfilDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio: %v", err)
	}

	// Generar nombre único con prefijo profile_
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("profile_%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(perfilDir, filename)

	// Guardar archivo
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo: %v", err)
	}
	defer dst.Close()

	// Posicionar el puntero al inicio
	file.Seek(0, 0)

	// Copiar contenido
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

// ============================================
// FUNCIONES AUXILIARES
// ============================================

// DeleteFile - Elimina un archivo del sistema
func DeleteFile(filePath string) error {
	// Convertir ruta web a ruta del sistema
	if strings.HasPrefix(filePath, "/static/") {
		filePath = strings.TrimPrefix(filePath, "/static/")
		filePath = "static/" + filePath
	}

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error eliminando archivo: %v", err)
	}
	return nil
}

// GetFileExtension - Obtiene la extensión de un archivo
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// IsValidImageExt - Verifica si la extensión es válida
func IsValidImageExt(filename string) bool {
	ext := GetFileExtension(filename)
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	return validExts[ext]
}

// GenerateUniqueFilename - Genera nombre único para archivo
func GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}

// GetMimeType - Obtiene el tipo MIME de un archivo
func GetMimeType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	file.Seek(0, 0)
	return http.DetectContentType(buffer), nil
}
