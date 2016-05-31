package spyfall

import (
	"testing"
	"encoding/json"
	"strings"
	"log"
	"time"
	"fmt"
)

type FakeConn struct {
	Out, In chan []byte
}

func NewFakeConn() *FakeConn {
	return &FakeConn{make(chan []byte), make(chan []byte)}
}

func (f *FakeConn) ReadJSON(data interface{}) error {
	bytes, ok := <-f.In
	if !ok {
		return fmt.Errorf("Channel closed")
	}
	return json.Unmarshal(bytes, data)
}

func (f *FakeConn) r(data interface{}) error {
	bytes := <-f.Out
	return json.Unmarshal(bytes, data)
}

func (f *FakeConn) WriteJSON(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Unable to send: %v", x)
		}
	}()
	f.Out <- bytes
	return err
}

func (f *FakeConn) w(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	f.In <- bytes
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

	log.Println("New")
	conn1.w(map[string]string{
		"action": "new",
		"name": "player1",
	})

	var data map[string]interface{}
	log.Println("Page")
	conn1.r(&data)

	if !strings.HasPrefix(data["page"].(string), "/game/") || data["type"].(string) != "page" {
		t.Fatal(data)
	}

	pageParts := strings.Split(data["page"].(string), "/")
	gameId := pageParts[len(pageParts)-1]

	log.Println("Game")
	conn1.r(&data)
	if data["type"].(string) != "game" {
		t.Fatal(data)
	}

	player2 := NewPlayer()
	conn2 := NewFakeConn()
	player2.Connect(conn2)
	go GameLoop(player2)

	log.Println("Join")
	conn2.w(map[string]string{
		"action": "join",
		"gameId": gameId,
		"name": "player2",
	})

	log.Println("Update")
	conn1.r(&data)

	if data["type"].(string) != "game" {
		if data["type"].(string) == "say" {
			t.Fatal(data["msg"])
		} else {
			t.Fatal(data)
		}
	}

	log.Println("Waiting")
	timeout := time.After(1 * time.Second)
	select {
	case <-conn1.In:
		t.Fatal("Unexpected message from conn1")
	case <-conn2.In:
		t.Fatal("Unexpected message from conn2")
	case <-timeout:
		// ok
	}

	log.Println("Done")

	close(conn1.In)
	close(conn2.In)
	close(conn1.Out)
	close(conn2.Out)
}
