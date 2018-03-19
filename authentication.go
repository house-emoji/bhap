package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// requireLogin is middleware that requires the user be logged in.
func requireLogin(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		loginSession, err := sessionStore.Get(r, "login")
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

func userFromSession(ctx context.Context, r *http.Request) (user, *datastore.Key, error) {
	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		return user{}, nil, fmt.Errorf("could not decode session: %v", err)
	}

	if loginSession.IsNew {
		return user{}, nil, errors.New("user is not logged in")
	}

	email := loginSession.Values["email"].(string)

	return userByEmail(ctx, email)
}

func deleteSession(w http.ResponseWriter, r *http.Request) error {
	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		return fmt.Errorf("could not decode session: %v", err)
	}

	loginSession.Options.MaxAge = -1
	loginSession.Save(r, w)

	return nil
}
