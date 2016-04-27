package spyfall

import (
	"net/http"
	"log"
	"math/rand"
	"strconv"
	"github.com/gorilla/websocket"
	"time"
)

var cookies *Cookies
var games *Games

func init() {
	rand.Seed(time.Now().Unix())
	cookies = NewCookies()
	games = NewGames()

	cookies.Load()
	games.Load()
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
	origin := r.Header.Get("Origin")
	if origin == "http://localhost:5000" || origin == "http://jakecoffman.com" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	cookie, err := r.Cookie("spyfall")
	switch {
	case err == nil:
		if cookies.Get(cookie.Value) != nil {
			break
		}
		log.Println("Database was reset?")
		cookies.Delete(cookie.Value)
		fallthrough
	case err == http.ErrNoCookie:
		player := NewPlayer("", nil)
		log.Println("New", player)
		cookie = &http.Cookie{Name: "spyfall", Value: player.Id}
		cookies.Set(player)
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
	player := cookies.Get(cookie.Value)
	if player == nil {
		log.Println("Player not created?", player)
		w.WriteHeader(500)
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
	player.conn = conn
	player.Connected = true

	game := games.FindGameByPlayerLeft(player)
	if game != nil {
		game.Rejoin(player)
	}

	GameLoop(conn, player, game)
}

func GameLoop(conn *websocket.Conn, player *Player, game *Game) {
	for {
		// wait for new/join game message
		msg := map[string]string{}
		if err := conn.ReadJSON(&msg); err != nil {
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
			player.Name = msg["name"]
			game = NewGame(gameId)
			games.Set(game)
			game.Join(player)
		case msg["action"] == "join":
			if game != nil {
				game.Leave(player)
			}
			gameId := msg["gameId"]
			player.Name = msg["name"]
			game = games.Get(gameId)
			if game == nil {
				log.Println("Game not found")
			} else {
				game.Join(player)
			}
		case msg["action"] == "rejoin":
			if game == nil {
				break
			}
			game.Rejoin(player)
		case msg["action"] == "start":
			if game == nil {
				break
			}
			game.Start()
		case msg["action"] == "end":
			if game == nil {
				break
			}
			game.End()
		case msg["action"] == "leave":
			if game != nil {
				game.Leave(player)
				game = nil
			}
			resp := map[string]string{}
			resp["type"] = "page"
			resp["page"] = "/"
			if err := conn.WriteJSON(resp); err != nil {
				log.Println(err)
			}
		default:
			log.Println("WAT:", msg)
		}
	}
}
