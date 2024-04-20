package internal

import (
	"math/rand"
	"strconv"
)

type Session struct {
	SessionId string
	MasterKey string
	Questions []Question
}

type Question struct {
	To   string
	Body string
}

func NewSession() Session {
	sessionId := ""
	for i := 0; i < 6; i++ {
		sessionId += strconv.Itoa(rand.Intn(9))
	}

	return Session{
		SessionId: sessionId,
		MasterKey: strconv.Itoa(rand.Int()),
	}
}
