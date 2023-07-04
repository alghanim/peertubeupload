package login

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"peertubeupload/model"
	"sync"
	"time"
)

var expirationTime time.Time

// var AccessToken model.AccessToken

type LoginManager struct {
	AccessToken model.AccessToken
	mutex       sync.Mutex
}

func (lm *LoginManager) LoginPrerequisite(baseURL string, client *http.Client) (*model.Login, error) {

	url := baseURL + "/oauth-clients/local"
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return &model.Login{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return &model.Login{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &model.Login{}, err
	}

	r, err := model.UnmarshalLogin(body)
	if err != nil {
		return &model.Login{}, err
	}

	return &r, nil

}

func (lm *LoginManager) Login(baseURL string, client *http.Client, loginClient *model.Login, grant_type string, username string, password string) error {

	apiurl := baseURL + "/users/token"
	method := "POST"
	data := url.Values{
		"client_id":     {loginClient.ClientID},
		"client_secret": {loginClient.ClientSecret},
		"grant_type":    {grant_type},
		"response_type": {"code"},
		"username":      {username},
		"password":      {password},
	}

	req, err := http.NewRequest(method, apiurl, bytes.NewBufferString(data.Encode()))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	accessToken, err := model.UnmarshalAccessToken(body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("not authorized")
	}

	lm.AccessToken = model.AccessToken{
		AccessToken:           accessToken.AccessToken,
		RefreshToken:          accessToken.RefreshToken,
		TokenType:             accessToken.TokenType,
		ExpiresIn:             accessToken.ExpiresIn,
		RefreshTokenExpiresIn: accessToken.RefreshTokenExpiresIn,
	}
	return nil

}

func (lm *LoginManager) UpdateTokenIfNeeded(baseURL string, client *http.Client, loginClient *model.Login, grant_type string, username string, password string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	// Check if the current time is after the token expiration time
	if time.Now().After(expirationTime) {
		// Get a new token
		var err error
		if lm.AccessToken.RefreshToken == "" {
			// If we don't have a refresh token, do a full login
			err = lm.Login(baseURL, client, loginClient, grant_type, username, password)
		} else {
			// If we have a refresh token, use it to get a new access token
			err = lm.RefreshAccessToken(baseURL, client, loginClient, lm.AccessToken.RefreshToken)
		}
		if err != nil {
			return err
		}

		// Set the new expiration time
		expirationTime = time.Now().Add(time.Second * time.Duration(lm.AccessToken.ExpiresIn))
	}
	return nil
}
func (lm *LoginManager) RefreshAccessToken(baseURL string, client *http.Client, loginClient *model.Login, refreshToken string) error {
	apiurl := baseURL + "/users/token"
	method := "POST"
	data := url.Values{
		"client_id":     {loginClient.ClientID},
		"client_secret": {loginClient.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequest(method, apiurl, bytes.NewBufferString(data.Encode()))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	accessToken, err := model.UnmarshalAccessToken(body)
	if err != nil {
		return err
	}

	lm.AccessToken = model.AccessToken{
		AccessToken:           accessToken.AccessToken,
		RefreshToken:          accessToken.RefreshToken,
		TokenType:             accessToken.TokenType,
		ExpiresIn:             accessToken.ExpiresIn,
		RefreshTokenExpiresIn: accessToken.RefreshTokenExpiresIn,
	}

	return nil
}

func (lm *LoginManager) GetAccessToken() string {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	return lm.AccessToken.AccessToken
}
