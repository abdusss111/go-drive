package e2e

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

func TestUserFullWorkflow(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Регистрация
	email := fmt.Sprintf("e2e_%d@example.com", time.Now().UnixNano())
	password := "password123"
	username := fmt.Sprintf("e2e_user_%d", time.Now().UnixNano())

	registerPayload := map[string]string{
		"email":    email,
		"password": password,
		"username": username,
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
		"email":    email,
		"password": password,
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
		"name":        "e2e-test-bucket",
		"description": "E2E test bucket",
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

	// 4. Загрузка 3 файлов
	fileNames := []string{"file1.txt", "file2.txt", "file3.txt"}
	fileContents := []string{"Content 1", "Content 2", "Content 3"}
	var fileIDs []string

	for i, name := range fileNames {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, _ := writer.CreateFormFile("file", name)
		part.Write([]byte(fileContents[i]))
		writer.Close()

		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/v1/buckets/%s/files", baseURL, bucketID), &buf)
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err = client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var fileResp struct {
			ID string `json:"id"`
		}
		body, _ = io.ReadAll(resp.Body)
		json.Unmarshal(body, &fileResp)
		resp.Body.Close()

		fileIDs = append(fileIDs, fileResp.ID)
	}

	// 5. Получить список файлов
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

	assert.Len(t, filesResp.Files, 3)

	// 6. Скачивание каждого файла
	for i, fileID := range fileIDs {
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/v1/buckets/%s/files/%s/download", baseURL, bucketID, fileID), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		resp, err = client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		content, _ := io.ReadAll(resp.Body)
		assert.Equal(t, fileContents[i], string(content))
		resp.Body.Close()
	}

	// 7. Удаление файлов
	for _, fileID := range fileIDs {
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/v1/buckets/%s/files/%s", baseURL, bucketID, fileID), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		resp, err = client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()
	}

	// 8. Удаление бакета
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/v1/buckets/%s", baseURL, bucketID), nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}
