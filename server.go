package spyfall

import (
	"net/http"
	"log"
	"math/rand"
	"strconv"
	"github.com/gorilla/websocket"
	"time"
	"strings"
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
	mux.HandleFunc("/register", Middleware(RegisterHandler))
	mux.HandleFunc("/ws", Middleware(WebsocketHandler))

	return mux
}

func Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r)
		log.Println(r.URL, r.Method)
	}
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/" + r.URL.Path)
}

// since gorilla doesn't let me set headers on the handshake...
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if strings.HasPrefix(origin, "http://localhost") || origin == "http://www.jakecoffman.com" {
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
		player := NewPlayer()
		log.Println("New", player)
		cookie = &http.Cookie{Name: "spyfall", Value: player.id}
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
	origin := r.Header.Get("Origin")
	if strings.HasPrefix(origin, "http://localhost") || origin == "http://www.jakecoffman.com" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
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
	player.Connect(conn)
	GameLoop(player)
}

func GameLoop(player *Player) {
	game := games.Rejoin(player)

	for {
		// react to messages from this player
		msg := map[string]string{}
		if err := player.conn.ReadJSON(&msg); err != nil {
			player.Disconnect()
			game.update()
			return
		}
		log.Println(msg)

		switch {
		case msg["action"] == "new":
			game.Leave(player)
			gameId := strconv.Itoa(rand.Int())[0:7]
			player.Page("/game/" + gameId)
			player.name = msg["name"]
			game = NewGame(gameId)
			games.Set(game)
			game.Join(player)
		case msg["action"] == "join":
			game.Leave(player)
			gameId := msg["gameId"]
			player.name = msg["name"]
			game = games.Get(gameId)
			if game == nil {
				player.Say("No such game")
			} else {
				game.Join(player)
			}
		case msg["action"] == "rejoin":
			game.TryRejoin(player)
		case msg["action"] == "start":
			game.Start(player)
		case msg["action"] == "end":
			game.End(player)
		case msg["action"] == "leave":
			game.Leave(player)
			game = nil
			player.Page("/")
		case msg["action"] == "kick":
			if game.Kick(player, msg["player"]) {
				player.Say("Player kicked")
			}
		default:
			response := "I don't understand '" + msg["action"] + "'"
			log.Println(response)
			player.Say(response)
		}
	}
}
