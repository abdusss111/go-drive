package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRegistrationToFileDownload(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Регистрация
	registerPayload := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
		"username": "testuser",
	}

	registerBody, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", baseURL+"/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 2. Логин
	loginPayload := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", baseURL+"/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &loginResp)
	resp.Body.Close()

	authToken := loginResp.Token
	require.NotEmpty(t, authToken)

	// 3. Создание бакета
	bucketPayload := map[string]interface{}{
		"name":        "test-bucket",
		"description": "Test bucket for integration",
	}

	bucketBody, _ := json.Marshal(bucketPayload)
	req, _ = http.NewRequest("POST", baseURL+"/v1/buckets", bytes.NewBuffer(bucketBody))
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var bucketResp struct {
		ID string `json:"id"`
	}
	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &bucketResp)
	resp.Body.Close()

	bucketID := bucketResp.ID
	require.NotEmpty(t, bucketID)

	// 4. Загрузка файла
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("Hello, this is test content!"))
	writer.Close()

	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/v1/buckets/%s/files", baseURL, bucketID), &buf)
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 5. Скачивание файла
	// Предположим, что ID файла возвращается при загрузке
	// Здесь нужно немного изменить логику, чтобы получить fileID
	// или использовать фикстуры для получения ID файла
	// или просто получить список файлов и выбрать первый

	// Получить список файлов
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/v1/buckets/%s/files", baseURL, bucketID), nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var filesResp struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &filesResp)
	resp.Body.Close()

	require.NotEmpty(t, filesResp.Files)
	fileID := filesResp.Files[0].ID

	// Скачивание
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/v1/buckets/%s/files/%s/download", baseURL, bucketID, fileID), nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	content, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Hello, this is test content!", string(content))
	resp.Body.Close()
}
