package spyfall

import (
	"sync"
	"fmt"
	"log"
	"math/rand"
)

type Game struct {
	sync.RWMutex
	Id       string `json:"gameId"`
	Players  []*Player `json:"players"`
	Left     []*Player `json:"disconnected"`
	Started  bool `json:"started"`
	Location string `json:"location"` // don't send to spy
	Info     map[string]*PlayerInfo `json:"playerInfo"`
}

func NewGame(gameId string) *Game {
	return &Game{
		Id: gameId,
		Players: []*Player{},
		Left: []*Player{},
		Info: map[string]*PlayerInfo{},
	}
}

type PlayerInfo struct {
	IsSpy bool // only sent to spy
	Role  string `json:"role"` // don't assign one to spy
}

func (g *Game) String() string {
	return fmt.Sprintf("Game: %v", g.Id)
}

func (g *Game) Join(player *Player) {
	g.Lock()
	g.Players = append(g.Players, player)
	log.Println(player, "joined", g)
	g.Unlock()
	g.update()
}

func (g *Game) Disconnect(player *Player) {
	g.Lock()
	for i, p := range g.Players {
		if p.Id == player.Id {
			g.Players = append(g.Players[:i], g.Players[i + 1:]...)
			break
		}
	}
	g.Left = append(g.Left, player)
	player.Connected = false
	g.Unlock()
	g.update()
}

func (g *Game) Leave(player *Player) {
	g.Lock()
	for i, p := range g.Players {
		if p.Id == player.Id {
			g.Players = append(g.Players[:i], g.Players[i + 1:]...)
			break
		}
	}
	log.Println(player, "left", g)
	g.Unlock()
	g.update()
}

func (g *Game) Rejoin(player *Player) {
	g.Lock()
	found := false
	for i, p := range g.Left {
		if p.Id == player.Id {
			g.Left = append(g.Left[:i], g.Left[i + 1:]...)
			found = true
			break
		}
	}
	if found {
		log.Println(player, "rejoined", g)
		g.Players = append(g.Players, player)
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
	g.Location = placeData.Locations[nLoc]
	roles := placeData.Roles[g.Location]
	nRole := rand.Intn(len(roles))
	nSpy := rand.Intn(len(g.Players) + len(g.Left))

	var i int
	var p *Player
	for i, p = range g.Players {
		pi := &PlayerInfo{}
		if i == nSpy {
			pi.IsSpy = true
		} else {
			pi.Role = roles[nRole]
			nRole += 1
			if nRole > len(roles) {
				nRole = 0
			}
		}
		g.Info[p.Id] = pi
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

	// TODO: on;y send player data of the player this is
	g.RLock()
	msg := map[string]interface{}{}
	msg["type"] = "game"
	msg["game"] = g
	for _, p := range g.Players {
		if err := p.conn.WriteJSON(msg); err != nil {
			log.Println(err, p.Id)
		}
	}
	g.RUnlock()
}
