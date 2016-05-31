package spyfall

import (
	"testing"
	"encoding/json"
	"strings"
	"log"
	"time"
)

type FakeConn struct {
	Chan chan []byte
}

func NewFakeConn() *FakeConn {
	return &FakeConn{make(chan []byte)}
}

func (f *FakeConn) ReadJSON(data interface{}) error {
	bytes := <-f.Chan
	return json.Unmarshal(bytes, data)
}

func (f *FakeConn) WriteJSON(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	f.Chan <- bytes
	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestIntegration(t *testing.T) {
	player1 := NewPlayer()
	conn1 := NewFakeConn()
	player1.Connect(conn1)
	go GameLoop(player1)

	conn1.WriteJSON(map[string]string{
		"action": "new",
		"name": "player1",
	})

	var data map[string]interface{}
	conn1.ReadJSON(&data)

	if !strings.HasPrefix(data["page"].(string), "/game/") || data["type"].(string) != "page" {
		t.Fatal(data)
	}

	pageParts := strings.Split(data["page"].(string), "/")
	gameId := pageParts[len(pageParts)-1]

	player2 := NewPlayer()
	conn2 := NewFakeConn()
	player2.Connect(conn2)
	go GameLoop(player2)

	conn2.WriteJSON(map[string]string{
		"action": "join",
		"gameId": gameId,
		"name": "player2",
	})

	conn1.ReadJSON(&data)
	conn2.ReadJSON(&data)

	if data["type"].(string) != "game" {
		if data["type"].(string) == "say" {
			t.Fatal(data["msg"])
		} else {
			t.Fatal(data)
		}
	}
	if data["type"].(string) != "game" {
		if data["type"].(string) == "say" {
			t.Fatal(data["msg"])
		} else {
			t.Fatal(data)
		}
	}

	timeout := time.After(1 * time.Second)
	select {
	case <-conn1.Chan:
		t.Fatal("Unexpected message")
	case <-conn2.Chan:
		t.Fatal("Unexpected message")
	case <-timeout:
		// ok
	}

	close(conn1.Chan)
	close(conn2.Chan)
}
