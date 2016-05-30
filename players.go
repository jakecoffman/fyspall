package spyfall

import (
	"os"
	"log"
	"encoding/json"
	"sync"
)

func NewCookies() *Cookies {
	return &Cookies{cookies: map[string]*Player{}}
}

type Cookies struct {
	sync.RWMutex
	cookies map[string]*Player
}

func (c *Cookies) Get(key string) *Player {
	c.RLock()
	defer c.RUnlock()
	return c.cookies[key]
}

func (c *Cookies) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.cookies, key)
}

func (c *Cookies) Set(player *Player) {
	c.Lock()
	c.cookies[player.id] = player
	c.Unlock()
	c.Save()
}

func (c *Cookies) Save() {
	c.RLock()
	defer c.RUnlock()
	file, err := os.Create("cookies.json")
	if err != nil {
		log.Println(err)
		return
	}
	if err = json.NewEncoder(file).Encode(c.cookies); err != nil {
		log.Println(err)
	}
}

func (c *Cookies) Load() error {
	c.Lock()
	defer c.Unlock()
	file, err := os.Open("cookies.json")
	if err != nil {
		log.Println(err)
		return err
	}
	if err = json.NewDecoder(file).Decode(&c.cookies); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
