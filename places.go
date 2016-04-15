package spyfall

import (
	"log"
	"encoding/json"
	"os"
)

var placeData Places

func init() {
	placeData.Load()
}

type Places struct{
	Locations []string
	Roles map[string][]string
}

func (l *Places) Load() {
	fp, err := os.Open("locations.json")
	if err != nil {
		log.Fatal(err)
	}
	l.Roles = map[string][]string{}
	if err = json.NewDecoder(fp).Decode(&l.Roles); err != nil {
		log.Fatal(err)
	}
	l.Locations = []string{}
	for k, _ := range l.Roles {
		l.Locations = append(l.Locations, k)
	}
}
