package main

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var listTemplate = compileTempl("views/list.html")

// listPageFiller fills the BHAP list page template.
type listPageFiller struct {
	LoggedIn bool
	Email    string
	Items    []listPageItemFiller
}

// listPageItemFiller fills a single BHAP in the BHAP list page template.
type listPageItemFiller struct {
	ID    int
	Title string
}

// serveListPage serves a page with a list of all BHAPs.
func serveListPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bhaps, err := allBHAPs(ctx)
	if err != nil {
		log.Errorf(ctx, "could not load BHAPs: %v", err)
		http.Error(w, "Failed to load BHAPs", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	currUser, userKey, err := userFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	filler := listPageFiller{
		LoggedIn: userKey != nil,
		Email:    currUser.Email,
		Items:    bhapsToListItemFillers(bhaps),
	}

	showTemplate(ctx, w, listTemplate, filler)
}

// bhapsToListItemFillers converts a BHAP to a list item filler, so it can be
// displayed in the BHAP list template.
func bhapsToListItemFillers(bhaps []bhap) []listPageItemFiller {
	filler := make([]listPageItemFiller, 0)
	for _, val := range bhaps {
		filler = append(filler, listPageItemFiller{
			ID:    val.ID,
			Title: val.Title,
		})
	}
	return filler
}
