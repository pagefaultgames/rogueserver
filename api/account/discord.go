/*
	Copyright (C) 2024 - 2026  Pagefault Games

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
	"fmt"
	"net/http"
	"net/url"

	"github.com/bwmarrin/discordgo"
)

var (
	DiscordClientID     string
	DiscordClientSecret string
	DiscordCallbackURL  string

	DiscordSession *discordgo.Session
	DiscordGuildID string
)

var (
	ErrEmptyDiscordId   = errors.New("discord returned empty user id")
	ErrEmptyAccessToken = errors.New("discord returned empty access token")
)

// DiscordProvider provides methods for handling Discord OAuth2 authentication.
type DiscordProvider interface {
	HandleDiscordCallback(r *http.Request) (string, error)
	RetrieveDiscordId(code string) (string, error)
	IsUserDiscordAdmin(discordId string, discordGuildID string) (bool, error)
}

type discordProvider struct {
	client   *http.Client
	tokenURL string
	userURL  string
}

var Discord = &discordProvider{
	client:   oauthHTTPClient,
	tokenURL: "https://discord.com/api/oauth2/token",
	userURL:  "https://discord.com/api/users/@me",
}

// Ensure that discordProvider implements DiscordProvider.
// This causes a compile-time error if the interface drifts,
// and has no runtime cost.
var _ DiscordProvider = (*discordProvider)(nil)

func (s *discordProvider) HandleDiscordCallback(r *http.Request) (string, error) {
	code, err := codeFromCallback(r)
	if err != nil {
		return "", err
	}

	discordId, err := s.RetrieveDiscordId(code)
	if err != nil {
		return "", err
	}

	return discordId, nil
}

func (s *discordProvider) RetrieveDiscordId(code string) (string, error) {
	v := make(url.Values)
	v.Set("client_id", DiscordClientID)
	v.Set("client_secret", DiscordClientSecret)
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", DiscordCallbackURL)
	v.Set("scope", "identify")

	token, err := s.client.PostForm(s.tokenURL, v)
	if err != nil {
		return "", err
	}

	defer token.Body.Close()

	if token.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discord token exchange failed: %s", token.Status)
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
		return "", ErrEmptyAccessToken
	}

	req, err := http.NewRequest("GET", s.userURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+access_token)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discord users/@me failed: %s", resp.Status)
	}

	type User struct {
		Id string `json:"id"`
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return "", err
	}

	if user.Id == "" {
		return "", ErrEmptyDiscordId
	}

	return user.Id, nil
}

func (s *discordProvider) IsUserDiscordAdmin(discordId string, discordGuildID string) (bool, error) {
	// fetch all roles from discord
	roles, err := DiscordSession.GuildRoles(discordGuildID)
	if err != nil {
		return false, err
	}

	// fetch all roles from user
	userRoles, err := DiscordSession.GuildMember(discordGuildID, discordId)
	if err != nil {
		return false, err
	}

	// check if user has a "Dev" or a "Division Heads" role
	var hasRole bool
	for _, role := range userRoles.Roles {
		for _, guildRole := range roles {
			if role == guildRole.ID && (guildRole.Name == "Dev" || guildRole.Name == "Division Heads" || guildRole.Name == "Helper") {
				hasRole = true
				break
			}
		}
	}

	if !hasRole {
		return false, nil
	}

	return true, nil
}
