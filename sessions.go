package bhap

import (
	"context"
	"fmt"
	"net/http"

	cascadestore "github.com/dsoprea/goappenginesessioncascade"
	"github.com/gorilla/sessions"
	"google.golang.org/appengine/datastore"
)

var sessionStore *cascadestore.CascadeStore

func init() {
	sessionStore = cascadestore.NewCascadeStore(
		cascadestore.DistributedBackends, []byte("23c124b173d"))
}

func GetSession(r *http.Request) (*sessions.Session, error) {
	return sessionStore.Get(r, "login")
}

// UserFromSession gets the currently logged in User based on session
// information. If no user is logged in, the returned key will be nil.
func UserFromSession(ctx context.Context, r *http.Request) (User, *datastore.Key, error) {
	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		return User{}, nil, fmt.Errorf("could not decode session: %v", err)
	}

	if loginSession.IsNew {
		return User{}, nil, nil
	}

	email := loginSession.Values["email"].(string)

	return UserByEmail(ctx, email)
}

// DeleteSession deletes the current session information.
func DeleteSession(w http.ResponseWriter, r *http.Request) error {
	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		return fmt.Errorf("could not decode session: %v", err)
	}

	loginSession.Options.MaxAge = -1
	loginSession.Save(r, w)

	return nil
}
