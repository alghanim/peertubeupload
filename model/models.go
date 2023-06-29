package model

import "encoding/json"

func UnmarshalLogin(data []byte) (Login, error) {
	var r Login
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Login) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Login struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func UnmarshalAccessToken(data []byte) (AccessToken, error) {
	var r AccessToken
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AccessToken) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type AccessToken struct {
	TokenType             string `json:"token_type"`
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	ExpiresIn             int64  `json:"expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

func UnmarshalVideo(data []byte) (Video, error) {
	var r Video
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Video) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Video struct {
	Video VideoClass `json:"video"`
}

type VideoClass struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	ShortUUID string `json:"shortUUID"`
}
