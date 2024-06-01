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
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	code := r.URL.Query().Get("code")
	gameUrl := os.Getenv("GAME_URL")
	if code == "" {
		defer http.Redirect(w, r, gameUrl, http.StatusSeeOther)
		return "", errors.New("code is empty")
	}

	googleId, err := RetrieveGoogleId(code)
	if err != nil {
		defer http.Redirect(w, r, gameUrl, http.StatusSeeOther)
		return "", err
	}

	return googleId, nil
}

func RetrieveGoogleId(code string) (string, error) {
	token, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"client_id":     {os.Getenv("GOOGLE_CLIENT_ID")},
		"client_secret": {os.Getenv("GOOGLE_CLIENT_SECRET")},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {os.Getenv("GOOGLE_CALLBACK_URI")},
	})

	if err != nil {
		return "", err
	}
	defer token.Body.Close()
	type TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		IdToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	var tokenResponse TokenResponse
	err = json.NewDecoder(token.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	userId, err := parseJWTWithoutValidation(tokenResponse.IdToken)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func parseJWTWithoutValidation(idToken string) (string, error) {
	parser := jwt.NewParser()

	// Use ParseUnverified to parse the token without validation
	parsedJwt, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, ok := parsedJwt.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return claims.GetSubject()
}
