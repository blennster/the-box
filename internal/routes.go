package internal

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

const (
	roomCookieName       = "room"
	masterRoomCookieName = "roomMasterKey"
)

var (
	Templates   *template.Template
	RoomStorage Storage[string, Room]
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	try(w, executeWithBase(w, "home.html", nil))
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	room := NewRoom()
	RoomStorage.Store(room.RoomId, room)

	http.SetCookie(w, &http.Cookie{
		Name:  roomCookieName,
		Value: room.RoomId,
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  masterRoomCookieName,
		Value: room.MasterKey,
		Path:  "/",
	})

	w.Header().Add("HX-Push-Url", fmt.Sprintf("/room/%s", room.RoomId))
	try(w, Templates.ExecuteTemplate(w, "master.html", room))
}

func HandleLeave(w http.ResponseWriter, r *http.Request) {
	roomCookie, err := r.Cookie(roomCookieName)
	if err != nil {
		fmt.Fprintf(w, "could not find room")
	}

	RoomStorage.Delete(roomCookie.Value)

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

	roomId, err := r.Cookie(roomCookieName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "there was no room cookie set")
		return
	}

	to := r.FormValue("to")
	body := r.FormValue("body")

	q := Question{
		To:   to,
		Body: body,
	}

	room, ok := RoomStorage.Load(roomId.Value)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "room %s was not found", roomId.Value)
		return
	}

	room.Questions = append(room.Questions, q)

	RoomStorage.Store(room.RoomId, room)
}

func HandleCount(w http.ResponseWriter, r *http.Request) {
	roomId, err := r.Cookie("room")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "there was no room cookie set")
		return
	}

	room, ok := RoomStorage.Load(roomId.Value)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "room %s was not found", roomId.Value)
		return
	}

	try(w, Templates.ExecuteTemplate(w, "count.html", room))
}

func HandleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "not a POST request")
		return
	}

	roomId := r.FormValue("roomid")

	_, ok := RoomStorage.Load(roomId)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "room %s was not found", roomId)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "room",
		Value: roomId,
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

	room, ok := getRoom(r, w)
	if !ok {
		return
	}

	type s struct {
		Question Question
		End      int
		Pos      int
	}
	var data s

	if len(room.Questions) >= n {
		action := r.URL.Query().Get("a")
		switch action {
		case "next":
			n = min(len(room.Questions)-1, n+1)
		case "prev":
			n = max(0, n-1)
		}

		data = s{
			Question: room.Questions[n],
			Pos:      n,
			End:      len(room.Questions),
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

	room, ok := RoomStorage.Load(roomId)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	roomCookie := http.Cookie{
		Name:  roomCookieName,
		Value: room.RoomId,
		Path:  "/",
	}
	http.SetCookie(w, &roomCookie)

	masterCookie, err := r.Cookie(masterRoomCookieName)
	if err == nil && masterCookie.Value == room.MasterKey {
		try(w, executeWithBase(w, "master.html", room))
		return
	}

	try(w, executeWithBase(w, "prompt.html", nil))
}

func HandleCheckOpen(w http.ResponseWriter, r *http.Request) {
	// We don't really care if the cookie is not found,
	// the same action should be taken
	roomId, _ := r.Cookie(roomCookieName)
	_, ok := RoomStorage.Load(roomId.Value)

	if !ok {
		try(w, Templates.ExecuteTemplate(w, "room_closed.html", nil))
	}
}

// Get the room from cookies
func getRoom(r *http.Request, w http.ResponseWriter) (Room, bool) {
	roomId, err := r.Cookie(roomCookieName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "there was no room cookie set")
		return Room{}, false
	}

	room, ok := RoomStorage.Load(roomId.Value)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "room %s was not found", roomId.Value)
		return Room{}, false
	}

	return room, true
}
