package spyfall

import (
	"testing"
	"encoding/json"
	"reflect"
)

func TestPlayer_JSON(t *testing.T) {
	p := NewPlayer()
	p.name = "Bob"
	p.connected = true
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Failed to marshal player")
	}
	p2 := &Player{}
	err = json.Unmarshal(data, p2)
	if err != nil {
		t.Fatalf("Failed to unmarshal player")
	}
	if !reflect.DeepEqual(p, p2) {
		t.Fatalf("original player differs from marshalled/unmarshalled:\n%#v\n%#v", p, p2)
	}
}
