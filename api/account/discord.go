/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

var (
	DiscordClientID     string
	DiscordClientSecret string
	DiscordCallbackURL  string
)

func HandleDiscordCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, GameURL, http.StatusSeeOther)
		return "", errors.New("code is empty")
	}

	discordId, err := RetrieveDiscordId(code)
	if err != nil {
		http.Redirect(w, r, GameURL, http.StatusSeeOther)
		return "", err
	}

	return discordId, nil
}

func RetrieveDiscordId(code string) (string, error) {
	v := make(url.Values)
	v.Set("client_id", DiscordClientID)
	v.Set("client_secret", DiscordClientSecret)
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", DiscordCallbackURL)
	v.Set("scope", "identify")

	token, err := http.PostForm("https://discord.com/api/oauth2/token", v)
	if err != nil {
		return "", err
	}

	// extract access_token from token
	type TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(token.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	access_token := tokenResponse.AccessToken
	if access_token == "" {
		return "", errors.New("access token is empty")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+access_token)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	type User struct {
		Id string `json:"id"`
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return "", err
	}

	return user.Id, nil
}

// TODO: fetch these instead of hardcoding them
var devsAndStaff = map[string]bool{}
var contributors = map[string]bool{}

func GetAccessGroupByDiscordRole(discordId string) (group string) {
	if devsAndStaff[discordId] {
		return DEV_STAFF
	}

	if contributors[discordId] {
		return CONTRIBUTOR
	}

	return ""
}
