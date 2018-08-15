package pages

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type bhapOperator struct {
	bhap    bhap.BHAP
	bhapKey *datastore.Key
	user    bhap.User
	userKey *datastore.Key
}

// bhapOperatorHandler is a handler that does some operation on a single BHAP.
// Middleware is provided here for convenience.
type bhapOperatorHandler func(op bhapOperator, w http.ResponseWriter, r *http.Request)

// SetUpBHAPOperator is middleware that fetches some common data required for
// BHAP operations.
func SetUpBHAPOperator(handler bhapOperatorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		// Load the BHAP
		loadedBHAP, bhapKey, err := bhapFromURLVars(ctx, mux.Vars(r))
		if err != nil {
			http.Error(w, "Could not load BHAP", http.StatusInternalServerError)
			log.Errorf(ctx, "loading BHAP: %v", err)
			return
		}
		if bhapKey == nil {
			http.Error(w, "No BHAP with that identifier", http.StatusNotFound)
			log.Warningf(ctx, "request for non-existent BHAP")
			return
		}

		user, userKey, err := bhap.UserFromSession(ctx, r)
		if err != nil {
			http.Error(w, "Could not load user", http.StatusInternalServerError)
			log.Errorf(ctx, "loading user: %v", err)
			return
		}
		if userKey == nil {
			http.Error(w, "You are not logged in", http.StatusForbidden)
			log.Warningf(ctx, "request from user that is not logged in")
			return
		}

		op := bhapOperator{
			bhap:    loadedBHAP,
			bhapKey: bhapKey,
			user:    user,
			userKey: userKey}

		handler(op, w, r)
	}
}
