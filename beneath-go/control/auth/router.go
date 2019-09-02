package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/core/httputil"

	"github.com/go-chi/chi"
	"github.com/markbates/goth/gothic"
)

// Router adds /github and /google login endpoints and /logout logout endpoint
func Router() http.Handler {
	// check config set
	if gothConfig == nil {
		log.Panic("Call InitGoth before AuthHandler")
		return nil
	}

	// prepare router
	router := chi.NewRouter()

	// social auth handlers
	router.With(providerParamToContext).MethodFunc("GET", "/{provider}", gothic.BeginAuthHandler)
	router.With(providerParamToContext).Method("GET", "/{provider}/callback", httputil.AppHandler(authCallbackHandler))

	// logout handler
	router.Method("GET", "/logout", httputil.AppHandler(logoutHandler))

	// done
	return router
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
func authCallbackHandler(w http.ResponseWriter, r *http.Request) error {
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

	// upsert user
	user, err := model.CreateOrUpdateUser(r.Context(), githubID, googleID, info.Email, info.Name, info.AvatarURL)
	if err != nil {
		return err
	}

	// create session secret
	secret, err := model.CreateUserSecret(r.Context(), user.UserID, model.SecretRoleManage, "Browser session")
	if err != nil {
		return err
	}

	// redirect to client, setting token
	url := fmt.Sprintf("%s/auth/callback/login?token=%s", gothConfig.ClientHost, url.QueryEscape(secret.SecretString))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	// done
	return nil
}

// logoutHandler revokes the current auth secret
func logoutHandler(w http.ResponseWriter, r *http.Request) error {
	secret := GetSecret(r.Context())
	if secret != nil {
		if secret.IsPersonal() {
			secret.Revoke(r.Context())
			log.Printf("Logout userID %s with hashed secret %s", secret.UserID, secret.HashedSecret)
		}
	}
	return nil
}
