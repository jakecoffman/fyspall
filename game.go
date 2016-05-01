package spyfall

import (
	"sync"
	"fmt"
	"log"
	"math/rand"
)

type Game struct {
	sync.RWMutex
	Id      string `json:"gameId"`
	Players map[string]*Player `json:"players"`
	Left    map[string]*Player `json:"disconnected"`
	Started bool `json:"started"`
	First   string `json:"first"`
	info    map[string]*PlayerInfo
}

func NewGame(gameId string) *Game {
	return &Game{
		Id: gameId,
		Players: map[string]*Player{},
		Left: map[string]*Player{},
		info: map[string]*PlayerInfo{},
	}
}

// This is private info, sent only to the players individualy
type PlayerInfo struct {
	// only sent to spy
	IsSpy    bool

	// not sent to spy
	Location string `json:"location"`
	Role     string `json:"role"`
}

func (g *Game) String() string {
	return fmt.Sprintf("Game: %v", g.Id)
}

func (g *Game) Join(player *Player) {
	g.Lock()
	g.Players[player.Id] = player
	log.Println(player, "joined", g)
	g.Unlock()
	g.update()
}

func (g *Game) Disconnect(player *Player) {
	g.Lock()
	delete(g.Players, player.Id)
	g.Left[player.Id] = player
	player.Connected = false
	g.Unlock()
	g.update()
}

func (g *Game) Leave(player *Player) {
	g.Lock()
	delete(g.Players, player.Id)
	log.Println(player, "left", g)
	g.Unlock()
	g.update()
}

// Or watch
func (g *Game) Rejoin(player *Player) {
	g.Lock()
	player, found := g.Left[player.Id]
	if found {
		delete(g.Left, player.Id)
		log.Println(player, "rejoined", g)
		g.Players[player.Id] = player
		g.Unlock()
		g.update()
	} else {
		log.Println(player.Id, "watching", g.Id)
		g.Unlock()
	}
}

func (g *Game) Start() {
	g.Lock()
	g.Started = true
	nLoc := rand.Intn(len(placeData.Locations))
	location := placeData.Locations[nLoc]
	roles := placeData.Roles[location]
	nRole := rand.Intn(len(roles))
	nSpy := rand.Intn(len(g.Players) + len(g.Left))
	nFirst := rand.Intn(len(g.Players) + len(g.Left))

	i := 0
	for id, _ := range g.Players {
		pi := &PlayerInfo{}
		if i == nFirst {
			g.First = id
		}
		if i == nSpy {
			pi.IsSpy = true
		} else {
			pi.Location = location
			pi.Role = roles[nRole]
			nRole += 1
			if nRole > len(roles) {
				nRole = 0
			}
		}
		g.info[id] = pi
		i += 1
	}
	g.Unlock()
	g.update()
}

func (g *Game) End() {
	g.Lock()
	g.Started = false
	g.Unlock()
	g.update()
}

func (g *Game) update() {
	// TODO: silly?
	games.Save()
	cookies.Save()

	g.RLock()
	msg := map[string]interface{}{}
	msg["type"] = "game"
	msg["game"] = g
	msg["info"] = placeData
	for id, p := range g.Players {
		msg["you"] = g.info[id]
		if err := p.conn.WriteJSON(msg); err != nil {
			log.Println(err, id)
		}
	}
	g.RUnlock()
}
