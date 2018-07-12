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
	r.HandleFunc("/bhap/{id}/edit", serveBHAPEditPage).
		Methods("GET")

	r.HandleFunc("/bhap/{id}/ready-for-discussion",
		setUpBHAPOperator(handleReadyForDiscussion)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/delete-vote",
		setUpBHAPOperator(handleDeleteVote)).
		Methods("GET")
	r.HandleFunc("/bhap/{id}/vote-accept",
		setUpBHAPOperator(handleVoteAccept)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/vote-reject",
		setUpBHAPOperator(handleVoteReject)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/withdraw",
		setUpBHAPOperator(handleWithdraw)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/edit",
		setUpBHAPOperator(handleEdit)).
		Methods("POST")
	// TODO(velovix): Implement replaced vote
	/*r.HandleFunc("/bhap/{id}/vote-replace").
	Methods("POST")*/

	r.Handle("/propose", requireLogin(serveNewBHAPPage)).
		Methods("GET")
	r.HandleFunc("/propose", handleNewBHAPForm).
		Methods("POST")

	r.HandleFunc("/login", serveLoginPage).
		Methods("GET")
	r.HandleFunc("/login", handleLoginForm).
		Methods("POST")
	r.Handle("/logout", requireLogin(logout)).
		Methods("GET")

	r.HandleFunc("/invite", serveInvitePage).
		Methods("GET")
	r.HandleFunc("/invite", handleInvitationForm).
		Methods("POST")

	r.HandleFunc("/new-user/{uid}", serveNewUserPage).
		Methods("GET")
	r.HandleFunc("/new-user/{uid}", handleNewUserForm).
		Methods("POST")

	r.HandleFunc("/tasks/send-invitations", sendInvitations)

	http.Handle("/", r)

	appengine.Main()
}
