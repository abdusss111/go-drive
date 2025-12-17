package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// CleanupUser удаляет пользователя и все его ресурсы
func CleanupUser(client *http.Client, authToken string) {
	// 1. Получить список бакетов
	req, _ := http.NewRequest("GET", baseURL+"/v1/buckets", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching buckets for cleanup: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var bucketsResp struct {
		Buckets []struct {
			ID string `json:"id"`
		} `json:"buckets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bucketsResp); err != nil {
		fmt.Printf("Error decoding buckets: %v\n", err)
		return
	}

	// 2. Удалить каждый бакет
	for _, bucket := range bucketsResp.Buckets {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/buckets/%s", baseURL, bucket.ID), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		client.Do(req)
	}
}

// SetupTestUser регистрирует нового пользователя для теста
func SetupTestUser(client *http.Client, email, password, username string) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
		"username": username,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to register user: %d", resp.StatusCode)
	}

	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}

	body, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", baseURL+"/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

// SetupTestUserWithCleanup оборачивает SetupTestUser и автоматически чистит после теста
func SetupTestUserWithCleanup(t *testing.T, client *http.Client) string {
	email := fmt.Sprintf("test_%s@example.com", uuid.NewString())
	username := fmt.Sprintf("test_%s", uuid.NewString())

	authToken, err := SetupTestUser(client, email, "password123", username)
	require.NoError(t, err)

	t.Cleanup(func() {
		CleanupUser(client, authToken)
	})

	return authToken
}
