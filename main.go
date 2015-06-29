package main

import (
	"encoding/json"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	hub = &Hub{
		Register:    make(chan *Connection),
		Unregister:  make(chan *Connection),
		Broadcast:   make(chan string),
		connections: make(map[*Connection]bool),
	}
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func serveWs(w http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			fmt.Println(err)
		}
		return
	}

	connection := &Connection{
		inputChannel: make(chan string),
		ws:           ws,
	}

	hub.Register <- connection

	go connection.StartWriter()
	connection.StartReader()
}

type TwitterConfig struct {
	ConsumerKey    string `json:"consumerKey"`
	ConsumerSecret string `json:"consumerSecret"`
	OauthToken     string `json:"oauthToken"`
	OauthSecret    string `json:"oauthSecret"`
}

func startTwitterStream() {
	data, err := ioutil.ReadFile("config.json")

	check(err)

	conf := &TwitterConfig{}
	err = json.Unmarshal(data, &conf)

	check(err)

	anaconda.SetConsumerKey(conf.ConsumerKey)
	anaconda.SetConsumerSecret(conf.ConsumerSecret)

	api := anaconda.NewTwitterApi(conf.OauthToken, conf.OauthSecret)

	find := strings.Join(os.Args[1:], ",")

	v := url.Values{}
	v.Set("track", find)

	stream := api.UserStream(v)

	for {
		select {
		case data := <-stream.C:
			t, ok := data.(anaconda.Tweet)

			if ok {
				select {
				case hub.Broadcast <- t.Text:
				default: // Do not block if hub isn't running
				}
			}
		}
	}
}

func main() {
	go hub.Run()
	go startTwitterStream()

	http.HandleFunc("/ws", serveWs)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.ListenAndServe(":8080", nil)
}
