package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AuthRequest struct {
	ClientID  string
	Username  string
	Password  string
	GrantType string
}

type AuthResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func main() {
	login("admin", "admin")
}

func login(username, password string) (*AuthResponse, error) {
	authReq := &AuthRequest{
		Username:  username,
		Password:  password,
		ClientID:  "admin-cli",
		GrantType: "password",
	}

	data := url.Values{}
	data.Set("client_id", authReq.ClientID)
	data.Set("username", authReq.Username)
	data.Set("password", authReq.Password)
	data.Set("grant_type", authReq.GrantType)

	req, err := http.NewRequest("POST", "http://localhost:8080/auth/realms/master/protocol/openid-connect/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := httpClient.Do(req)
	statusCode := resp.StatusCode
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode > 299 {
		return nil, errors.New("[ERROR]")
	}

	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		log.Printf("[ERROR] [%s]\n", err)
		return nil, err
	}

	return &authResponse, err
}

func request(method, baseURI, uri, contentType, token string, payload io.Reader) (int, []byte, error) {
	url := fmt.Sprintf("%s%s", baseURI, uri)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("%s%s", "Bearer ", token))
	}

	resp, err := httpClient.Do(req)
	statusCode := resp.StatusCode
	if err != nil {
		return statusCode, nil, err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if statusCode < 200 || statusCode > 299 {
		return statusCode, nil, errors.New("[ERROR]")
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	return statusCode, bytes, err
}
