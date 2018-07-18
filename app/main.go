package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/house-emoji/bhap/email"
	"github.com/house-emoji/bhap/pages"
	"google.golang.org/appengine"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", pages.ServeListPage)
	r.HandleFunc("/bhap", pages.ServeListPage)
	r.HandleFunc("/bhap/{id}", pages.ServeBHAPPage).
		Methods("GET")
	r.HandleFunc("/bhap/{id}/edit", pages.ServeEditBHAPPage).
		Methods("GET")

	r.HandleFunc("/bhap/{id}/ready-for-discussion",
		pages.SetUpBHAPOperator(pages.HandleReadyForDiscussion)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/delete-vote",
		pages.SetUpBHAPOperator(pages.HandleDeleteVote)).
		Methods("GET")
	r.HandleFunc("/bhap/{id}/vote-accept",
		pages.SetUpBHAPOperator(pages.HandleVoteAccept)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/vote-reject",
		pages.SetUpBHAPOperator(pages.HandleVoteReject)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/withdraw",
		pages.SetUpBHAPOperator(pages.HandleWithdraw)).
		Methods("POST")
	r.HandleFunc("/bhap/{id}/edit",
		pages.SetUpBHAPOperator(pages.HandleEdit)).
		Methods("POST")
	// TODO(velovix): Implement replaced vote
	/*r.HandleFunc("/bhap/{id}/vote-replace").
	Methods("POST")*/

	r.Handle("/propose", pages.RequireLogin(pages.ServeNewBHAPPage)).
		Methods("GET")
	r.HandleFunc("/propose", pages.HandleNewBHAPForm).
		Methods("POST")

	r.HandleFunc("/login", pages.ServeLoginPage).
		Methods("GET")
	r.HandleFunc("/login", pages.HandleLoginForm).
		Methods("POST")
	r.Handle("/logout", pages.RequireLogin(pages.HandleLogoutForm)).
		Methods("GET")

	r.HandleFunc("/invite", pages.ServeInvitePage).
		Methods("GET")
	r.HandleFunc("/invite", pages.HandleInvitationForm).
		Methods("POST")

	r.HandleFunc("/new-user/{uid}", pages.ServeNewUserPage).
		Methods("GET")
	r.HandleFunc("/new-user/{uid}", pages.HandleNewUserForm).
		Methods("POST")

	r.HandleFunc("/tasks/send-invitations", email.SendInvitations)

	http.Handle("/", r)

	appengine.Main()
}
