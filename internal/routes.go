package internal

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

const (
	sessionCookieName       = "session"
	masterSessionCookieName = "sessionMasterKey"
)

var (
	Templates      *template.Template
	SessionStorage Storage[string, Session]
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	try(w, executeWithBase(w, "home.html", nil))
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	session := NewSession()
	SessionStorage.Store(session.SessionId, session)

	http.SetCookie(w, &http.Cookie{
		Name:  sessionCookieName,
		Value: session.SessionId,
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  masterSessionCookieName,
		Value: session.MasterKey,
		Path:  "/",
	})

	w.Header().Add("HX-Push-Url", fmt.Sprintf("/room/%s", session.SessionId))
	try(w, Templates.ExecuteTemplate(w, "master.html", session))
}

func HandleLeave(w http.ResponseWriter, r *http.Request) {
	sessCookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		fmt.Fprintf(w, "could not find session")
	}

	SessionStorage.Delete(sessCookie.Value)

	w.Header().Add("HX-Push-Url", "/")
	try(w, Templates.ExecuteTemplate(w, "home.html", nil))
}

func HandleAddQuestion(w http.ResponseWriter, r *http.Request) {
	if try(w, Templates.ExecuteTemplate(w, "prompt.html", nil)) != nil {
		return
	}

	if r.Method != "POST" {
		return
	}

	sessionId, err := r.Cookie(sessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "there was no session cookie set")
		return
	}

	to := r.FormValue("to")
	body := r.FormValue("body")

	q := Question{
		To:   to,
		Body: body,
	}

	sess, ok := SessionStorage.Load(sessionId.Value)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "session %s was not found", sessionId.Value)
		return
	}

	sess.Questions = append(sess.Questions, q)

	SessionStorage.Store(sess.SessionId, sess)
}

func HandleCount(w http.ResponseWriter, r *http.Request) {
	sessionId, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "there was no session cookie set")
		return
	}

	sess, ok := SessionStorage.Load(sessionId.Value)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "session %s was not found", sessionId.Value)
		return
	}

	try(w, Templates.ExecuteTemplate(w, "count.html", sess))
}

func HandleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "not a POST request")
		return
	}

	sessionId := r.FormValue("sessionid")

	_, ok := SessionStorage.Load(sessionId)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "session %s was not found", sessionId)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: sessionId,
	})

	if err := Templates.ExecuteTemplate(w, "prompt.html", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error occured executing template (%s)", err)
		return
	}
}

func HandleView(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		fmt.Fprintf(w, "no index was supplied")
		return
	}

	sess, ok := getSession(r, w)
	if !ok {
		return
	}

	type s struct {
		Question Question
		End      int
		Pos      int
		Next     int
		Prev     int
	}

	data := s{}
	if len(sess.Questions) >= n {
		data = s{
			Question: sess.Questions[n],
			Pos:      n + 1,
			End:      len(sess.Questions),
			Next:     min(len(sess.Questions), n+1),
			Prev:     max(0, n-1),
		}
	}

	if err := Templates.ExecuteTemplate(w, "view.html", data); err != nil {
		fmt.Fprintf(w, "error occured executing template (%s)", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func HandleRoom(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomId")

	sess, ok := SessionStorage.Load(roomId)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sessionCookie := http.Cookie{
		Name:  sessionCookieName,
		Value: sess.SessionId,
		Path:  "/",
	}
	http.SetCookie(w, &sessionCookie)

	masterCookie, err := r.Cookie(masterSessionCookieName)
	if err == nil && masterCookie.Value == sess.MasterKey {
		try(w, executeWithBase(w, "master.html", sess))
		return
	}

	try(w, executeWithBase(w, "prompt.html", nil))
}

func getSession(r *http.Request, w http.ResponseWriter) (Session, bool) {
	sessionId, err := r.Cookie("session")
	if err != nil {
		fmt.Fprintf(w, "there was no session cookie set")
		w.WriteHeader(http.StatusBadRequest)
		return Session{}, false
	}

	sess, ok := SessionStorage.Load(sessionId.Value)
	if !ok {
		fmt.Fprintf(w, "session %s was not found", sessionId.Value)
		// w.WriteHeader(http.StatusBadRequest)
		return Session{}, false
	}

	return sess, true
}
