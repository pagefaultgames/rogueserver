package api

import (
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
)

// TODO: actual randomized key for sessions
const (
	key    = "randomString"
	MaxAge = 86400 * 30 // 30 days
	IsProd = false
)

func InitAuth(discordClientId string, discordClientSecret string, discordCallbackURI string) {
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd

	gothic.Store = store

	goth.UseProviders(
		discord.New(discordClientId, discordClientSecret, discordCallbackURI, discord.ScopeIdentify),
	)
}
