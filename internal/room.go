package internal

import (
	"math/rand"
	"strconv"
)

type Room struct {
	RoomId    string
	MasterKey string
	Questions []Question
}

type Question struct {
	To   string
	Body string
}

func NewRoom() Room {
	roomId := ""
	for i := 0; i < 6; i++ {
		roomId += strconv.Itoa(rand.Intn(9))
	}

	return Room{
		RoomId:    roomId,
		MasterKey: strconv.Itoa(rand.Int()),
	}
}
