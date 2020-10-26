package control

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// AuthOptions configures Beneath authentication
type AuthOptions struct {
	Github struct {
		ID     string `mapstructure:"id"`
		Secret string `mapstructure:"secret"`
	}
	Google struct {
		ID     string `mapstructure:"id"`
		Secret string `mapstructure:"secret"`
	}
}

func (s *Server) initGoth() {
	// add providers
	goth.UseProviders(
		google.New(
			s.Options.Auth.Google.ID, s.Options.Auth.Google.Secret, s.authCallbackURL("google"),
			"https://www.googleapis.com/auth/plus.login", "email",
		),
		github.New(
			s.Options.Auth.Github.ID, s.Options.Auth.Github.Secret, s.authCallbackURL("github"),
			"user:email",
		),
	)

	// set store for oauth state
	store := sessions.NewCookieStore([]byte(s.Options.SessionSecret))
	store.Options.HttpOnly = true
	store.Options.Secure = strings.HasPrefix(s.Options.Host, "https")
	gothic.Store = store
}

func (s *Server) authCallbackURL(provider string) string {
	return fmt.Sprintf("%s/auth/%s/callback", s.Options.Host, provider)
}

func (s *Server) registerAuth() {
	// social auth handlers
	router := s.Router.With(providerParamToContext)
	router.MethodFunc("GET", "/auth/{provider}", gothic.BeginAuthHandler)
	router.Method("GET", "/auth/{provider}/callback", httputil.AppHandler(s.authCallbackHandler))

	// logout handler
	router.Method("GET", "/auth/logout", httputil.AppHandler(s.logoutHandler))
}

// providerParamToContext is a middleware that reads the url param "provider" and
// saves it in the request context -- necessary because goth reads the provider name
// from the key "provider" in the request context
func providerParamToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), interface{}("provider"), provider)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authCallbackHandler gets called after social authentication
func (s *Server) authCallbackHandler(w http.ResponseWriter, r *http.Request) error {
	// handle with gothic
	info, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return err
	}

	// we're not using gothic for auth management, so end the user session immediately
	gothic.Logout(w, r)

	// get googleID or githubID
	var googleID string
	var githubID string
	if info.Provider == "google" {
		googleID = info.UserID
	} else if info.Provider == "github" {
		githubID = info.UserID
	} else {
		return httputil.NewError(500, "expected provider to be 'google' or 'github'")
	}

	// we only want to use Github's nicknames
	var nickname string
	if githubID != "" {
		nickname = info.NickName
	}

	// upsert user
	user, err := s.Users.CreateOrUpdateUser(r.Context(), githubID, googleID, info.Email, nickname, info.Name, info.AvatarURL)
	if err != nil {
		return err
	}

	// create session secret
	secret, err := s.Secrets.CreateUserSecret(r.Context(), user.UserID, "Browser session", false, false)
	if err != nil {
		return err
	}

	// redirect to client, setting token
	url := fmt.Sprintf("%s/-/redirects/auth/login/callback?token=%s", s.Options.FrontendHost, url.QueryEscape(secret.Token.String()))
	http.Redirect(w, r, url, http.StatusSeeOther)

	// done
	return nil
}

// logoutHandler revokes the current auth secret
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) error {
	secret := middleware.GetSecret(r.Context())
	if secret != nil {
		if secret.IsUser() {
			s.Secrets.RevokeUserSecret(r.Context(), secret.(*models.UserSecret))
			log.S.Infow(
				"control user logout",
				"user_id", secret.GetOwnerID(),
				"secret_id", secret.GetSecretID(),
			)
		}
	}
	return nil
}
