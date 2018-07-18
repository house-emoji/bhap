package pages

import (
	"net/http"

	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// RequireLogin is middleware that requires the user be logged in.
func RequireLogin(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		loginSession, err := bhap.GetSession(r)
		if err != nil {
			http.Error(w, "Could not decode session", http.StatusInternalServerError)
			log.Errorf(ctx, "could not decode session: %v", err)
			return
		}

		if loginSession.IsNew {
			log.Infof(ctx, "No session exists, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r)
	})
}
