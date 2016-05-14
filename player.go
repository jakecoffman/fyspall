package spyfall

import (
	"strconv"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
)

func NewPlayer(name string, conn *websocket.Conn) *Player {
	return &Player{
		Id: strconv.Itoa(rand.Int()), // TODO: UUID
		Name: name,
		conn: conn,
		Connected: false,
	}
}

type Player struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	conn      *websocket.Conn
	Connected bool `json:"connected"`
}

func (p *Player) String() string {
	return fmt.Sprintf("Player: %v", p.Id)
}
