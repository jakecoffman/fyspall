package spyfall

import (
	"testing"
	"encoding/json"
	"strings"
	"log"
)

type FakeConn struct {
	Data chan []byte
}

func NewFakeConn() *FakeConn {
	return &FakeConn{make(chan []byte)}
}

func (f *FakeConn) ReadJSON(data interface{}) error {
	return json.Unmarshal(<-f.Data, data)
}

func (f *FakeConn) WriteJSON(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	f.Data <- bytes
	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestIntegration(t *testing.T) {
	player1 := NewPlayer()
	conn1 := NewFakeConn()
	defer close(conn1.Data)

	player1.Connect(conn1)
	go GameLoop(player1)

	conn1.WriteJSON(map[string]string{
		"action": "new",
		"name": "Bob",
	})

	var data map[string]string
	conn1.ReadJSON(&data)

	if !strings.HasPrefix(data["page"], "/game/") || data["type"] != "page" {
		t.Fatal("unexpected", data)
	}
}
