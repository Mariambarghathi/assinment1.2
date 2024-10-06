package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func JasonResponseHandler(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func HandleErrors(w http.ResponseWriter, status int, message string) {
	JasonResponseHandler(w, status, map[string]string{
		"error": message,
	})
}

func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SaveImageFile saves the uploaded image file to a specified directory with a new name
func SaveImageFile(file io.Reader, table string, filename string) (string, error) {
	// Create directory structure if it doesn't exist
	fullPath := filepath.Join("uploads", table)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return "", err
	}

	// Generate new filename
	randomNumber := rand.Intn(1000)
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	newFileName := fmt.Sprintf("%s_%d_%d%s", filepath.Base(table), timestamp, randomNumber, ext)
	newFilePath := filepath.Join(fullPath, newFileName)

	// Save the file
	destFile, err := os.Create(newFilePath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		return "", err
	}

	// Return the full path including directory
	return newFilePath, nil
}
