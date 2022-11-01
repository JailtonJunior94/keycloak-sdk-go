package keycloak

import (
	"context"
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

type KeycloakSDK struct {
	BaseURL    string
	Username   string
	Password   string
	Context    context.Context
	Session    *AuthResponse
	HTTPClient *http.Client
}

func NewKeycloakSDK(ctx context.Context, baseURL, username, password string) (*KeycloakSDK, error) {
	keycloakSDK := &KeycloakSDK{
		BaseURL:    baseURL,
		Username:   username,
		Password:   password,
		Context:    ctx,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}

	session, err := keycloakSDK.auth(baseURL, username, password)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	keycloakSDK.Session = session

	return keycloakSDK, nil
}

func (k *KeycloakSDK) auth(baseURL, username, password string) (*AuthResponse, error) {
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

	uri := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", baseURL)
	req, err := http.NewRequestWithContext(k.Context, http.MethodPost, uri, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Println(err)
		return nil, errors.New("[ERROR]")
	}

	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		log.Printf("[ERROR] [%s]\n", err)
		return nil, err
	}

	return &authResponse, nil
}

func (k *KeycloakSDK) request(method, baseURI, uri, contentType, token string, payload io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", baseURI, uri)
	req, err := http.NewRequestWithContext(k.Context, method, url, payload)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("%s%s", "Bearer ", token))
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(fmt.Sprintf("[ERROR] [StatusCode] [%d] [Detail] [%s]", resp.StatusCode, string(b)))
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	return bytes, err
}
