package spyfall

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Games struct {
	sync.RWMutex
	games map[string]*Game
}

func NewGames() *Games {
	return &Games{games: map[string]*Game{}}
}

func (g *Games) Get(gameId string) *Game {
	g.RLock()
	defer g.RUnlock()
	return g.games[gameId]
}

func (g *Games) Set(game *Game) {
	g.Lock()
	g.games[game.id] = game
	g.Unlock()
	g.Save()
}

func (g *Games) Save() {
	g.RLock()
	defer g.RUnlock()
	file, err := os.Create("games.json")
	if err != nil {
		log.Println(err)
		return
	}
	if err = json.NewEncoder(file).Encode(g.games); err != nil {
		log.Println(err)
	}
}

func (g *Games) Load() error {
	g.Lock()
	defer g.Unlock()
	file, err := os.Open("games.json")
	if err != nil {
		log.Println(err)
		return err
	}
	if err = json.NewDecoder(file).Decode(&g.games); err != nil {
		log.Println(err)
		return err
	}
	// mark all players as disconnected since the server restarted
	for _, game := range g.games {
		for _, p := range game.players {
			p.connected = false
		}
	}
	return nil
}

func (g *Games) Rejoin(player *Player) *Game {
	g.RLock()
	for _, game := range g.games {
		if ok := game.TryRejoin(player); ok {
			log.Println(player, "rejoined", game)
			g.RUnlock()
			game.update()
			return game
		}
	}
	g.RUnlock()
	return nil
}
