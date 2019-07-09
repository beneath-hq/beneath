package auth

import (
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

// GothConfig used to setup https://github.com/markbates/goth
type GothConfig struct {
	ClientHost       string
	BackendHost      string
	SessionSecret    string
	GithubAuthID     string
	GithubAuthSecret string
	GoogleAuthID     string
	GoogleAuthSecret string
}

var (
	gothConfig *GothConfig
)

// InitGoth configures goth
func InitGoth(config *GothConfig) {
	if config == nil {
		log.Panicf("GothConfig cannot be nil")
		return
	}

	if gothConfig != nil {
		log.Panicf("InitGoth called twice")
		return
	}

	gothConfig = config

	goth.UseProviders(
		google.New(
			config.GoogleAuthID, config.GoogleAuthSecret, callbackURL("google"),
			"https://www.googleapis.com/auth/plus.login", "email",
		),
		github.New(
			config.GithubAuthID, config.GithubAuthSecret, callbackURL("github"),
			"user:email",
		),
	)

	// set store for oauth state
	store := sessions.NewCookieStore([]byte(config.SessionSecret))
	store.Options.HttpOnly = true
	store.Options.Secure = strings.HasPrefix(config.BackendHost, "https")
	gothic.Store = store
}

func callbackURL(provider string) string {
	return fmt.Sprintf("%s/auth/%s/callback", gothConfig.BackendHost, provider)
}
