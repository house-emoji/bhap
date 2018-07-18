package email

import "html/template"

// compileTempl wraps the common template compiling pattern. Panics in case of
// error.
func compileTempl(filename string) *template.Template {
	return template.Must(template.ParseFiles(filename))
}
