package spyfall

import (
	"net/http"
	"log"
	"math/rand"
	"strconv"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var cookies map[string]*Player
var games map[string]*Game

func init() {
	rand.Seed(time.Now().Unix())
	cookies = map[string]*Player{}
	games = map[string]*Game{}
}

func NewServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", RootHandler)
	mux.HandleFunc("/register", RegisterHandler)
	mux.HandleFunc("/ws", WebsocketHandler)

	return mux
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/" + r.URL.Path)
}

// since gorilla doesn't let me set headers on the handshake...
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	cookie, err := r.Cookie("spyfall")
	switch {
	case err == nil:
		log.Println("Already have cookie:", cookie.Value)
	case err == http.ErrNoCookie:
		val := strconv.Itoa(rand.Int())
		log.Println("Setting new cookie", val)
		cookie = &http.Cookie{Name: "spyfall", Value: val}
		cookies[val] = nil
		http.SetCookie(w, cookie)
	default:
		log.Println(err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	cookie, err := r.Cookie("spyfall")
	if err != nil {
		log.Println("No cookie", cookie, err)
		w.WriteHeader(400)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	player := cookies[cookie.Value]
	var game *Game

	if player != nil {
		log.Println("Player already exists", player.id)
		game = FindGameByPlayerLeft(player)
		if game != nil {
			log.Println("Player is in game", game.GameId)
			game.Rejoin(player)
		}
	} else {
		player = NewPlayer("", conn)
		cookies[cookie.Value] = player
	}

	GameLoop(conn, player, game)
}

func FindGameByPlayerLeft(player *Player) *Game {
	for _, game := range games {
		for _, p := range game.left {
			if player.id == p.id {
				p.conn = player.conn
				return game
			}
		}
	}
	return nil
}

func GameLoop(conn *websocket.Conn, player *Player, game *Game) {
	for {
		// wait for new/join game message
		msg := map[string]string{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println(err)
			if game != nil {
				game.Disconnect(player)
			}
			return
		}

		switch {
		case msg["action"] == "new":
			if game != nil {
				game.Leave(player)
			}
			gameId := strconv.Itoa(rand.Int())
			resp := map[string]string{}
			resp["type"] = "page"
			resp["page"] = "/game/" + gameId
			if err := conn.WriteJSON(resp); err != nil {
				log.Println(err)
			}
			player := NewPlayer(msg["name"], conn)
			game = NewGame(gameId)
			games[gameId] = game
			log.Println(games)
			game.Join(player)
		case msg["action"] == "join":
			if game != nil {
				game.Leave(player)
			}
			gameId := msg["gameId"]
			player.Name = msg["name"]
			log.Println(player.Name, "joining", gameId)
			game = games[gameId]
			log.Println(games)
			if game == nil {
				log.Println("Game not found")
			} else {
				game.Join(player)
			}
		case msg["action"] == "rejoin":
			game.update()
		default:
			log.Println("WAT:", msg)
		}
	}
}

func NewGame(gameId string) *Game {
	return &Game{GameId: gameId, Players: []*Player{}, left: []*Player{}}
}

func NewPlayer(name string, conn *websocket.Conn) *Player {
	return &Player{id: rand.Int(), Name: name, conn: conn}
}

type Player struct {
	id int
	Name string `json:"name"`
	conn *websocket.Conn
}

type Game struct {
	sync.RWMutex
	GameId string `json:"gameId"`
	Players []*Player `json:"players"`
	left []*Player // players that disconnected
}

func (g *Game) Join(player *Player) {
	g.Lock()
	g.Players = append(g.Players, player)
	log.Println(player.id, "joined", g.GameId)
	g.Unlock()
	g.update()
}

func (g *Game) Disconnect(player *Player) {
	g.Lock()
	for i, p := range g.Players {
		if p.id == player.id {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			break
		}
	}
	g.left = append(g.left, player)
	log.Println(player.id, "disconnect", g.GameId)
	g.Unlock()
	g.update()
}

func (g *Game) Leave(player *Player) {
	g.Lock()
	for i, p := range g.Players {
		if p.id == player.id {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			break
		}
	}
	log.Println(player.id, "left", g.GameId)
	g.Unlock()
	g.update()
}

func (g *Game) Rejoin(player *Player) {
	g.Lock()
	for i, p := range g.left {
		if p.id == player.id {
			g.left = append(g.left[:i], g.left[i+1:]...)
			break
		}
	}
	g.Players = append(g.Players, player)
	log.Println(player.id, "rejoined", g.GameId)
	g.Unlock()
	g.update()
}

func (g *Game) update() {
	g.RLock()
	msg := map[string]interface{}{}
	msg["type"] = "game"
	msg["game"] = g
	for _, p := range g.Players {
		if err := p.conn.WriteJSON(msg); err != nil {
			log.Println(err, g.GameId)
		}
	}
	g.RUnlock()
}
