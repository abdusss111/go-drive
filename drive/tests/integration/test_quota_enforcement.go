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

func TestQuotaEnforcement(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Логин
	loginPayload := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginPayload)
	req, _ := http.NewRequest("POST", baseURL+"/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &loginResp)
	resp.Body.Close()

	authToken := loginResp.Token

	// Создать бакет
	bucketPayload := map[string]interface{}{
		"name": "quota-test-bucket",
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

	// Загрузить файл, превышающий квоту (предполагаем, что квота = 1MB)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "large-file.bin")
	// Записать 2MB данных
	for i := 0; i < 2*1024*1024; i++ {
		_, _ = part.Write([]byte("x")) // Исправлено: используем _, _ чтобы избежать ошибки
	}
	writer.Close()

	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/v1/buckets/%s/files", baseURL, bucketID), &buf)
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err = client.Do(req)
	require.NoError(t, err)

	// Ожидаем ошибку, если квота включена
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}
