package spyfall

import (
	"strconv"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"sync"
	"log"
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
	sync.RWMutex
	Id        string `json:"id"`
	Name      string `json:"name"`
	conn      *websocket.Conn
	Connected bool `json:"connected"`
}

func (p *Player) String() string {
	p.RLock()
	defer p.RUnlock()
	return fmt.Sprintf("Player: %v", p.Id)
}

func (p *Player) Disconnect() {
	p.Lock()
	p.conn = nil
	p.Connected = false
	p.Unlock()
}

func (p *Player) WriteJSON(msg interface{}) {
	p.Lock()
	defer p.Unlock()
	if !p.Connected {
		return
	}
	if err := p.conn.WriteJSON(msg); err != nil {
		log.Println(err, p.Id)
		p.conn = nil
		p.Connected = false
	}
}
