package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPresignedURLFlow(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Логин
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

	// 2. Получить бакет (или создать)
	// Здесь предполагаем, что бакет уже создан
	// или создаем его заново
	// ...

	// 3. Загрузить файл
	// (аналогично предыдущему тесту)
	// ...

	// 4. Предположим, что у нас есть bucketID и fileID
	// Проверим, есть ли эндпоинт для генерации presigned URL
	// Если такого эндпоинта нет, то этот тест не имеет смысла
	// Пока пропускаем, если в API его нет
	t.Skip("Presigned URL endpoint not implemented in API yet")
}
