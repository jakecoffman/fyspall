package spyfall

import (
	"strconv"
	"fmt"
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
	conn      Connection
	connected bool
}

type Connection interface {
	WriteJSON(interface{}) error
	ReadJSON(interface{}) error
}

func (p Player) String() string {
	return fmt.Sprintf("Player: %v", p.id)
}

type jsonPlayer struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Connected bool `json:"connected"`
}

func (p Player) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonPlayer{
		Id: p.id,
		Name: p.name,
		Connected: p.connected,
	})
}

func (p *Player) UnmarshalJSON(b []byte) error {
	temp := &jsonPlayer{}
	err := json.Unmarshal(b, temp)
	p.id = temp.Id
	p.connected = temp.Connected
	p.name = temp.Name
	return err
}

func (p *Player) Connect(ws Connection) {
	p.Lock()
	defer p.Unlock()
	p.conn = ws
	p.connected = true
	log.Println(p, "has connected")
}

func (p *Player) Disconnect() {
	p.Lock()
	defer p.Unlock()
	p.conn = nil
	p.connected = false
	log.Println(p, "has disconnected")
}

func (p *Player) WriteJSON(msg interface{}) error {
	p.Lock()
	defer p.Unlock()
	if !p.connected {
		return nil
	}
	if err := p.conn.WriteJSON(msg); err != nil {
		log.Println("Player", p.id, "is disconnected")
		return err
	}
	return nil
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
