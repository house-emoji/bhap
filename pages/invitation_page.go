package pages

import (
	"net/http"

	"github.com/house-emoji/bhap"
	"github.com/rs/xid"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// ServeInvitePage serves the page that is used to create new invitations to
// join the BHAP consortium.
func ServeInvitePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/invite.html")
}

// HandleInvitationForm creates a new invitation based on form input from a
// POST request.
func HandleInvitationForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	email := r.FormValue("email")

	duplicateCount, err := datastore.NewQuery(bhap.UserEntityName).
		Filter("Email =", email).
		Count(ctx)
	if err != nil {
		log.Errorf(ctx, "looking for duplicate errors: %v", err)
		http.Error(w, "Could not look for duplicate emails",
			http.StatusInternalServerError)
		return
	}

	if duplicateCount > 0 {
		log.Warningf(ctx, "attempt to add a duplicate email invitation")
		http.Error(w, "A user with that email already exists",
			http.StatusBadRequest)
		return
	}

	newInvitation := bhap.Invitation{
		Email:     email,
		UID:       xid.New().String(),
		EmailSent: false,
	}

	key := datastore.NewKey(ctx, bhap.InvitationEntityName, "", 0, nil)
	if _, err := datastore.Put(ctx, key, &newInvitation); err != nil {
		log.Errorf(ctx, "could not create invitation: %v", err)
		http.Error(w, "Could not create invitation",
			http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "created a new invitation for %v", email)

	http.Redirect(w, r, "/invite", http.StatusSeeOther)
}
