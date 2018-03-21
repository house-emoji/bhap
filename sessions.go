package main

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/appengine/datastore"
)

// userFromSession gets the currently logged in user based on session
// information. If no user is logged in, the returned key will be nil.
func userFromSession(ctx context.Context, r *http.Request) (user, *datastore.Key, error) {
	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		return user{}, nil, fmt.Errorf("could not decode session: %v", err)
	}

	if loginSession.IsNew {
		return user{}, nil, nil
	}

	email := loginSession.Values["email"].(string)

	return userByEmail(ctx, email)
}

// deleteSession deletes the current session information.
func deleteSession(w http.ResponseWriter, r *http.Request) error {
	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		return fmt.Errorf("could not decode session: %v", err)
	}

	loginSession.Options.MaxAge = -1
	loginSession.Save(r, w)

	return nil
}
