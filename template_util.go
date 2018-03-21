package main

import (
	"context"
	"html/template"
	"net/http"

	"google.golang.org/appengine/log"
)

// compileTempl wraps the common template compiling pattern. Panics in case of
// error.
func compileTempl(filename string) *template.Template {
	return template.Must(template.ParseFiles(filename))
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
