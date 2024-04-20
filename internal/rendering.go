package internal

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
)

// Execute a template and wrap it with "base.html"
func executeWithBase(w io.Writer, name string, data any) error {
	return executeWith(w, "base.html", name, data)
}

// Execute a template and wrap it with some other template
func executeWith(w io.Writer, base string, name string, data any) error {
	var s strings.Builder
	if name != "" {
		if err := Templates.ExecuteTemplate(&s, name, data); err != nil {
			return err
		}
	}

	if err := Templates.ExecuteTemplate(w, base, template.HTML(s.String())); err != nil {
		// if err := Views["index"].Execute(w, nil); err != nil {
		return err
	}

	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func try(w http.ResponseWriter, err error) error {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error %s occured\n", err)
	}

	return err
}
