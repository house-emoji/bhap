package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

const dateFormat = "2006-01-02"

var (
	listTemplate    *template.Template
	bhapTemplate    *template.Template
	proposeTemplate *template.Template
	newUserTemplate *template.Template
)

func init() {
	listTemplate = template.Must(template.ParseFiles("views/list.html"))
	bhapTemplate = template.Must(template.ParseFiles("views/bhap.html"))
	proposeTemplate = template.Must(template.ParseFiles("views/propose.html"))
	newUserTemplate = template.Must(template.ParseFiles("views/new-user.html"))
}

// serveBHAPPage serves up a page that displays info on a single BHAP.
func serveBHAPPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the requested ID
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warningf(ctx, "invalid ID %v: %v", idStr, err)
		http.Error(w, "ID must be an integer", 400)
		return
	}

	// Load the requested BHAP
	loadedBHAP, exists, err := bhapByID(ctx, id)
	if err != nil {
		log.Errorf(ctx, "could not load BHAP: %v", err)
		http.Error(w, "Failed to load BHAP", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, fmt.Sprintf("No BHAP with ID %v", id), 404)
		log.Warningf(ctx, "unknown BHAP requested: %v", id)
		return
	}

	// Render the BHAP content
	html := string(blackfriday.Run([]byte(loadedBHAP.Content)))

	var author user
	if err := datastore.Get(ctx, loadedBHAP.Author, &author); err != nil {
		log.Errorf(ctx, "Error loading user: %v", err)
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	filler := bhapFiller{
		PaddedID:     fmt.Sprintf("%04d", loadedBHAP.ID),
		Title:        loadedBHAP.Title,
		LastModified: loadedBHAP.LastModified.Format(dateFormat),
		Author:       author.String(),
		Status:       loadedBHAP.Status,
		CreatedDate:  loadedBHAP.CreatedDate.Format(dateFormat),
		HTMLContent:  template.HTML(html),
	}
	showTemplate(ctx, w, bhapTemplate, filler)
}

// serveProposePage serves a page for creating new BHAPs.
func serveProposePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/propose.html")
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

	filler := bhapsToListItemFillers(bhaps)

	showTemplate(ctx, w, listTemplate, filler)
}

// showTemplate executes the given template with the given filler. If there's
// an error, an internal server error is reported.
func showTemplate(
	ctx context.Context,
	w http.ResponseWriter,
	templ *template.Template,
	filler interface{}) {

	if err := templ.Execute(w, filler); err != nil {
		http.Error(w,
			"Could not execute template",
			http.StatusInternalServerError)
		log.Errorf(ctx, "error executing template %v: %v",
			templ.Name(), err)
	}
}
