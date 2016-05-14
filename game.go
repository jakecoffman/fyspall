package spyfall

import (
	"sync"
	"fmt"
	"log"
	"math/rand"
	"time"
	"encoding/json"
)

type Game struct {
	sync.RWMutex
	id       string
	players  map[string]*Player
	started  bool
	creator  string
	first    string
	deadline time.Time
	info     map[string]*PlayerInfo
}

type jsonGame struct {
	Id       string `json:"gameId"`
	Players  map[string]*Player `json:"players"`
	Started  bool `json:"started"`
	Creator  string `json:"creator"`
	First    string `json:"first"`
	Deadline time.Time `json:"deadline"`
	Info     map[string]*PlayerInfo `json:"playerInfo"`
}

// This is private info, sent only to the players individually
type PlayerInfo struct {
	// only sent to spy
	IsSpy    bool

	// not sent to spy
	Location string `json:"location"`
	Role     string `json:"role"`
}

func NewGame(gameId string) *Game {
	return &Game{
		id: gameId,
		players: map[string]*Player{},
		info: map[string]*PlayerInfo{},
	}
}

func (g *Game) String() string {
	if g == nil {
		return "Game:<nil>"
	}
	return fmt.Sprintf("Game: %v", g.id)
}

func (g *Game) MarshalJSON() ([]byte, error) {
	g.RLock()
	defer g.RUnlock()
	return json.Marshal(&jsonGame{
		Id: g.id,
		Players: g.players,
		Started: g.started,
		Creator: g.creator,
		First: g.first,
		Deadline: g.deadline,
		Info: g.info,
	})
}

func (g *Game) UnmarshalJSON(b []byte) error {
	g.Lock()
	defer g.Unlock()
	jsonGame := &jsonGame{}
	err := json.Unmarshal(b, jsonGame)
	g.id = jsonGame.Id
	g.players = jsonGame.Players
	g.started = jsonGame.Started
	g.creator = jsonGame.Creator
	g.first = jsonGame.First
	g.deadline = jsonGame.Deadline
	g.info = jsonGame.Info
	return err
}

func (g *Game) Join(player *Player) {
	if g == nil {
		return
	}
	g.Lock()
	g.players[player.id] = player
	if len(g.players) == 1 {
		g.creator = player.id
	}
	log.Println(player, "joined", g)
	g.Unlock()
	g.update()
}

func (g *Game) Leave(player *Player) {
	if g == nil {
		return
	}
	g.Lock()
	delete(g.players, player.id)
	log.Println(player, "left", g)
	if g.creator == player.id && len(g.players) > 0 {
		nPlayer := rand.Intn(len(g.players))
		n := 0
		for id, _ := range g.players {
			if nPlayer == n {
				g.creator = id
				break
			}
			n += 1
		}
	}
	g.Unlock()
	g.update()
}

func (g *Game) TryRejoin(player *Player) bool {
	if g == nil {
		return false
	}
	g.RLock()
	if _, ok := g.players[player.id]; !ok {
		g.RUnlock()
		return false
	}
	g.RUnlock()
	g.Lock()
	g.players[player.id] = player
	g.Unlock()
	g.update()
	return true
}

func (g *Game) Start(p *Player) {
	if g == nil {
		return
	}
	g.RLock()
	if g.creator != p.id {
		p.Say("Only game creator can start the game")
		g.RUnlock()
		return
	}
	g.RUnlock()
	g.Lock()
	g.started = true
	nLoc := rand.Intn(len(placeData.Locations))
	location := placeData.Locations[nLoc]
	roles := placeData.Roles[location]
	nRole := rand.Intn(len(roles))
	nSpy := rand.Intn(len(g.players))
	nFirst := rand.Intn(len(g.players))

	i := 0
	for id, _ := range g.players {
		pi := &PlayerInfo{}
		if i == nFirst {
			g.first = id
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
	g.deadline = time.Now().Add(8 * time.Minute)
	g.Unlock()
	g.update()
}

func (g *Game) End(p *Player) {
	if g == nil {
		return
	}
	g.RLock()
	if g.creator != p.id {
		p.Say("Only game creator may end the game")
		g.RUnlock()
		return
	}
	g.RUnlock()
	g.Lock()
	g.started = false
	g.info = map[string]*PlayerInfo{}
	g.first = ""
	g.deadline = time.Time{}
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
		Id: g.id,
		Players: g.players,
		Started: g.started,
		Creator: g.creator,
		First: g.first,
		Deadline: g.deadline,
	}
	msg["info"] = placeData
	for id, p := range g.players {
		info := g.info[id]
		if info == nil {
			msg["you"] = struct{ Player *Player `json:"player"` }{p}
		} else {
			msg["you"] = EnhancedPlayerInfo{
				Player: p,
				IsSpy: info.IsSpy,
				Location: info.Location,
				Role: info.Role,
			}
		}
		p.WriteJSON(msg)
	}
}

type GameMessage struct {
	Id       string `json:"gameId"`
	Players  map[string]*Player `json:"players"`
	Left     map[string]*Player `json:"disconnected"`
	Started  bool `json:"started"`
	Creator  string `json:"creator"`
	First    string `json:"first"`
	Deadline time.Time `json:"deadline"`
}

type EnhancedPlayerInfo struct {
	Player   *Player `json:"player"`
	// only sent to spy
	IsSpy    bool `json:"isSpy"`

	// not sent to spy
	Location string `json:"location"`
	Role     string `json:"role"`
}