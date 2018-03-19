package main

import (
	"net/http"

	cascadestore "github.com/dsoprea/goappenginesessioncascade"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

var sessionStore *cascadestore.CascadeStore

func init() {
	sessionStore = cascadestore.NewCascadeStore(
		cascadestore.DistributedBackends, []byte("23c124b173d"))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", serveListPage)
	r.HandleFunc("/bhap", serveListPage)
	r.HandleFunc("/bhap/{id}", serveBHAPPage).
		Methods("GET")

	r.Handle("/propose", requireLogin(serveProposePage)).
		Methods("GET")
	r.HandleFunc("/propose", createBHAP).
		Methods("POST")

	r.HandleFunc("/login", serveLoginPage).
		Methods("GET")
	r.HandleFunc("/login", login).
		Methods("POST")
	r.Handle("/logout", requireLogin(logout)).
		Methods("GET")

	r.HandleFunc("/invite", serveInvitePage).
		Methods("GET")
	r.HandleFunc("/invite", createInvitation).
		Methods("POST")

	r.HandleFunc("/new-user/{uid}", serveNewUserPage).
		Methods("GET")
	r.HandleFunc("/new-user/{uid}", newUser).
		Methods("POST")

	r.HandleFunc("/tasks/send-invitations", sendInvitations)

	http.Handle("/", r)

	appengine.Main()
}
