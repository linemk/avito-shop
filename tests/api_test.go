package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

// AuthResponse структура ответа при аутентификации
type AuthResponse struct {
	Token string `json:"token"`
}

// SendCoinRequest структура запроса на отправку монет
type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

// InfoResponse – структура ответа от /api/info
type InfoResponse struct {
	Coins       int `json:"coins"`
	CoinHistory struct {
		Received []struct {
			Amount int `json:"amount"`
		} `json:"received"`
		Sent []struct {
			Amount int `json:"amount"`
		} `json:"sent"`
	} `json:"coinHistory"`
}

func authenticateUser(t *testing.T, username, password string) string {
	reqBody := []byte(`{"username": "` + username + `", "password": "` + password + `"}`)
	resp, err := http.Post(baseURL+"/api/auth", "application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err, "Auth request should not error")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for valid auth")

	var authResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	assert.NoError(t, err, "Decoding auth response should succeed")
	assert.NotEmpty(t, authResp.Token, "Token should not be empty")
	return authResp.Token
}

// сценарий с успешной аутентификацией пользователя
func TestAuth(t *testing.T) {
	token := authenticateUser(t, "testuser@gmail.com", "testpass")
	assert.NotEmpty(t, token, "token should be obtained")
}

// сценарий с безуспешной аутентификацией пользователя
func TestAuthInvalid(t *testing.T) {
	reqBody := []byte(`{"username": "", "password": ""}`)
	resp, err := http.Post(baseURL+"/api/auth", "application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 for invalid auth")
}

// сценарий с получением info
func TestGetInfo(t *testing.T) {
	token := authenticateUser(t, "infouser@test.com", "testpass")
	req, err := http.NewRequest("GET", baseURL+"/api/info", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK for /api/info")
}

// сценарий с получением info (пользователь не авторизован)
func TestGetInfoUnauthorized(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/api/info", nil)
	assert.NoError(t, err)
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 unauthorized for missing token")
}

// сценарий передачи монеток другим сотрудникам
func TestSendCoin(t *testing.T) {
	token := authenticateUser(t, "sender@test.com", "testpass")
	_ = authenticateUser(t, "receiver@test.com", "testpass")

	requestBody := SendCoinRequest{ToUser: "receiver@test.com", Amount: 10}
	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/sendCoin", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 for valid coin transfer")
}

// сценарий безуспешной передачи монеток другим сотрудникам
func TestSendCoinInvalid(t *testing.T) {
	token := authenticateUser(t, "sender@test.com", "testpass")

	requestBody := SendCoinRequest{ToUser: "", Amount: -5}
	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/sendCoin", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 for invalid coin transfer data")
}

// сценарий покупки мерча
func TestBuyItem(t *testing.T) {
	token := authenticateUser(t, "buyer@test.com", "testpass")
	item := "t-shirt"
	req, err := http.NewRequest("GET", baseURL+"/api/buy/"+item, nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 for buying an item")
}

// сценарий покупки мерча (с другим названием)
func TestBuyItemNotFound(t *testing.T) {
	token := authenticateUser(t, "buyer@test.com", "testpass")
	item := "nonexistent_item"
	req, err := http.NewRequest("GET", baseURL+"/api/buy/"+item, nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 for nonexistent item")
}

func TestSendCoinSelfTransfer(t *testing.T) {
	// Сначала аутентифицируем пользователя
	reqBody := []byte(`{"username": "selftransfer@test.com", "password": "testpass"}`)
	resp, err := http.Post(baseURL+"/api/auth", "application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var authResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, authResp.Token)

	sendReq := SendCoinRequest{ToUser: "selftransfer@test.com", Amount: 50}
	jsonBody, err := json.Marshal(sendReq)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/sendCoin", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "self-transfer should not be allowed")
}

// TestCoinHistoryVerification проверяет, что после перевода монет история транзакций формируется корректно.
func TestCoinHistoryVerification(t *testing.T) {
	// Аутентифицируем пользователя A
	tokenA := authenticateUser(t, "userA@test.com", "testpass")
	// Аутентифицируем пользователя B
	tokenB := authenticateUser(t, "userB@test.com", "testpass")

	// Пользователь A переводит 100 монет пользователю B
	requestBody := SendCoinRequest{ToUser: "userB@test.com", Amount: 100}
	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/sendCoin", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for coin transfer")

	// Проверяем историю для пользователя A (отправленных транзакций)
	reqInfoA, err := http.NewRequest("GET", baseURL+"/api/info", nil)
	assert.NoError(t, err)
	reqInfoA.Header.Set("Authorization", "Bearer "+tokenA)
	respInfoA, err := client.Do(reqInfoA)
	assert.NoError(t, err)
	defer respInfoA.Body.Close()

	var infoRespA InfoResponse
	err = json.NewDecoder(respInfoA.Body).Decode(&infoRespA)
	assert.NoError(t, err)
	var foundSent bool
	for _, sent := range infoRespA.CoinHistory.Sent {
		if sent.Amount == 100 {
			foundSent = true
			break
		}
	}
	assert.True(t, foundSent, "user A should have a sent transaction of 100 coins")

	// Проверяем историю для пользователя B (полученных транзакций)
	reqInfoB, err := http.NewRequest("GET", baseURL+"/api/info", nil)
	assert.NoError(t, err)
	reqInfoB.Header.Set("Authorization", "Bearer "+tokenB)
	respInfoB, err := client.Do(reqInfoB)
	assert.NoError(t, err)
	defer respInfoB.Body.Close()

	var infoRespB InfoResponse
	err = json.NewDecoder(respInfoB.Body).Decode(&infoRespB)
	assert.NoError(t, err)
	var foundReceived bool
	for _, received := range infoRespB.CoinHistory.Received {
		if received.Amount == 100 {
			foundReceived = true
			break
		}
	}
	assert.True(t, foundReceived, "user B should have a received transaction of 100 coins")
}
