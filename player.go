package spyfall

import (
	"strconv"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"sync"
	"log"
	"encoding/json"
)

func NewPlayer() *Player {
	return &Player{
		id: strconv.Itoa(rand.Int()), // TODO: UUID
	}
}

type Player struct {
	sync.RWMutex
	id        string
	name      string
	conn      *websocket.Conn
	connected bool
}

func (p *Player) String() string {
	p.RLock()
	defer p.RUnlock()
	return fmt.Sprintf("Player: %v", p.id)
}

type jsonPlayer struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Connected bool `json:"connected"`
}

func (p *Player) MarshalJSON() ([]byte, error) {
	p.RLock()
	defer p.RUnlock()
	return json.Marshal(&jsonPlayer{
		Id: p.id,
		Name: p.name,
		Connected: p.connected,
	})
}

func (p *Player) UnmarshalJSON(b []byte) error {
	p.RLock()
	defer p.RUnlock()
	temp := &jsonPlayer{}
	err := json.Unmarshal(b, temp)
	p.id = temp.Id
	p.connected = temp.Connected
	p.name = temp.Name
	return err
}

func (p *Player) Connect(ws *websocket.Conn) {
	p.Lock()
	p.conn = ws
	p.connected = true
	p.Unlock()
}

func (p *Player) Disconnect() {
	p.Lock()
	p.conn = nil
	p.connected = false
	p.Unlock()
}

func (p *Player) WriteJSON(msg interface{}) {
	p.RLock()
	defer p.RUnlock()
	if !p.connected {
		return
	}
	if err := p.conn.WriteJSON(msg); err != nil {
		log.Println("Player", p.id, "is disconnected")
		p.conn = nil
		p.connected = false
	}
}

func (p *Player) Page(page string) {
	p.WriteJSON(map[string]string{
		"type": "page",
		"page": page,
	})
}

func (p *Player) Say(msg string) {
	p.WriteJSON(map[string]string{
		"type": "say",
		"msg": msg,
	})
}
