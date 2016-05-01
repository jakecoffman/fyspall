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
	g.games[game.Id] = game
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
	// mark all players as "left" since the server restarted
	for _, game := range g.games {
		for id, p := range game.Players {
			game.Left[id] = p
		}
		game.Players = map[string]*Player{}
		for _, p := range game.Left {
			p.Connected = false
		}
	}
	return nil
}

func (g *Games) FindGameByPlayerLeft(player *Player) *Game {
	for _, game := range g.games {
		for _, p := range game.Left {
			if player.Id == p.Id {
				return game
			}
		}
	}
	return nil
}
