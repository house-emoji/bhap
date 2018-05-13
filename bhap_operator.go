package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type bhapOperator struct {
	bhap    bhap
	bhapKey *datastore.Key
	user    user
	userKey *datastore.Key
}

type bhapOperatorHandler func(op bhapOperator, w http.ResponseWriter, r *http.Request)

func setUpBHAPOperator(handler bhapOperatorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		bhapID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, "ID is not a string", http.StatusBadRequest)
			log.Warningf(ctx, "BHAP not an ID")
			return
		}

		bhap, key, err := bhapByID(ctx, bhapID)
		if err != nil {
			http.Error(w, "Could not load BHAP", http.StatusInternalServerError)
			log.Errorf(ctx, "loading BHAP: %v", err)
			return
		}
		if key == nil {
			http.Error(w, "No BHAP with ID", http.StatusNotFound)
			log.Warningf(ctx, "request for non-existent BHAP %v", bhapID)
			return
		}

		user, userKey, err := userFromSession(ctx, r)
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
			bhap:    bhap,
			bhapKey: key,
			user:    user,
			userKey: userKey}

		handler(op, w, r)
	}
}
