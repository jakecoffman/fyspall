package spyfall

import (
	"sync"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Game struct {
	sync.RWMutex
	Id       string `json:"gameId"`
	Players  map[string]*Player `json:"players"`
	Started  bool `json:"started"`
	First    string `json:"first"`
	Deadline time.Time `json:"deadline"`
	Info     map[string]*PlayerInfo `json:"playerInfo"`
}

func NewGame(gameId string) *Game {
	return &Game{
		Id: gameId,
		Players: map[string]*Player{},
		Info: map[string]*PlayerInfo{},
	}
}

// This is private info, sent only to the players individually
type PlayerInfo struct {
	// only sent to spy
	IsSpy    bool

	// not sent to spy
	Location string `json:"location"`
	Role     string `json:"role"`
}

func (g *Game) String() string {
	if g == nil {
		return "Game:<nil>"
	}
	return fmt.Sprintf("Game: %v", g.Id)
}

func (g *Game) Join(player *Player) {
	if g == nil {
		return
	}
	g.Lock()
	g.Players[player.Id] = player
	log.Println(player, "joined", g)
	g.Unlock()
	g.update()
}

func (g *Game) Leave(player *Player) {
	if g == nil {
		return
	}
	g.Lock()
	delete(g.Players, player.Id)
	log.Println(player, "left", g)
	g.Unlock()
	g.update()
}

func (g *Game) TryRejoin(player *Player) bool {
	if g == nil {
		return false
	}
	g.RLock()
	if _, ok := g.Players[player.Id]; !ok {
		g.RUnlock()
		return false
	}
	g.RUnlock()
	g.Lock()
	g.Players[player.Id] = player
	g.Unlock()
	g.update()
	return true
}

func (g *Game) Start() {
	if g == nil {
		return
	}
	g.Lock()
	g.Started = true
	nLoc := rand.Intn(len(placeData.Locations))
	location := placeData.Locations[nLoc]
	roles := placeData.Roles[location]
	nRole := rand.Intn(len(roles))
	nSpy := rand.Intn(len(g.Players))
	nFirst := rand.Intn(len(g.Players))

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
		g.Info[id] = pi
		i += 1
	}
	g.Deadline = time.Now().Add(8*time.Minute)
	g.Unlock()
	g.update()
}

func (g *Game) End() {
	if g == nil {
		return
	}
	g.Lock()
	g.Started = false
	g.Info = map[string]*PlayerInfo{}
	g.First = ""
	g.Deadline = time.Time{}
	g.Unlock()
	g.update()
}

func (g *Game) update() {
	if g == nil {
		return
	}
	g.Lock()
	games.Save()
	cookies.Save()
	g.Unlock()

	g.RLock()
	defer g.RUnlock()
	msg := map[string]interface{}{}
	msg["type"] = "game"
	msg["game"] = GameMessage{
		Id: g.Id,
		Players: g.Players,
		Started: g.Started,
		First: g.First,
		Deadline: g.Deadline,
	}
	msg["info"] = placeData
	for id, p := range g.Players {
		msg["you"] = g.Info[id]
		p.WriteJSON(msg)
	}
}

type GameMessage struct {
	Id       string `json:"gameId"`
	Players  map[string]*Player `json:"players"`
	Left     map[string]*Player `json:"disconnected"`
	Started  bool `json:"started"`
	First    string `json:"first"`
	Deadline time.Time `json:"deadline"`
}
