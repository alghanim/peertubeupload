package auth

import (
	"net/http"
	"peertubeupload/model"
)

type Authenticator interface {
	LoginPrerequisite(baseURL string, client *http.Client) (*model.Login, error)
	Login(baseURL string, client *http.Client, loginClient *model.Login, grant_type string, username string, password string) error
	UpdateTokenIfNeeded(baseURL string, client *http.Client, loginClient *model.Login, grant_type string, username string, password string) error
	RefreshAccessToken(baseURL string, client *http.Client, loginClient *model.Login, refreshToken string) error
	GetAccessToken() string
}
