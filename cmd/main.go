package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/blennster/the-box/internal"
	"github.com/blennster/the-box/web"
)

func main() {
	fs := web.FS
	tmpl := template.Must(template.ParseFS(fs, "*.html"))
	fmt.Println(tmpl.DefinedTemplates())
	internal.Templates = tmpl

	store := internal.NewInMemoryStore[string, internal.Session]()
	testSession := internal.NewSession()
	testSession.SessionId = "000000"
	store.Store(testSession.SessionId, testSession)

	internal.SessionStorage = &store

	http.Handle("/static/", http.FileServerFS(web.Static))
	http.HandleFunc("/", internal.HandleHome)
	http.HandleFunc("/create", internal.HandleCreate)
	http.HandleFunc("/leave", internal.HandleLeave)
	http.HandleFunc("/addquestion", internal.HandleAddQuestion)
	http.HandleFunc("/count", internal.HandleCount)
	http.HandleFunc("/join", internal.HandleJoin)
	http.HandleFunc("/view", internal.HandleView)
	http.HandleFunc("/room/{roomId}", internal.HandleRoom)

	addr := ":8000"
	fmt.Printf("listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("could not start server: %s\n", err)
	}
}
