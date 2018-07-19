package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	nats "github.com/nats-io/go-nats"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients = make(map[*websocket.Conn]bool)
	send    = make(chan *frontendMsg)
	receive = make(chan *frontendMsg)
	c       *nats.EncodedConn
)

type frontendMsg struct {
	Msg_type     int `json:"Msg_type"`
	ID           int `json:"ID"`
	Lat          int `json:"Lat"`
	Lng          int `json:"Lng"`
	CauseCode    int `json:"CauseCode"`
	SubCauseCode int `json:"SubCauseCode"`
	Timestamp    int `json:"Timestamp"`
	Termination  int `json:"Termination"`
}

func main() {
	port:= os.Getenv("PORT")
	log.Println(os.Getenv("NATS") + ":"+port)
	nc, err := nats.Connect("nats://" + os.Getenv("NATS")+ ":"+port)
	check(err)
	c, err = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	check(err)
	defer c.Close()
	c.Subscribe("msg.frontend", func(p *frontendMsg) {
		fmt.Println(p)
		send <- p
	})

	http.HandleFunc("/", handler)
	go sendSocketMessages()
	go sendNatsMessages()

	log.Println("http server started on :4000")
	err = http.ListenAndServe(":4000", nil)
	check(err)

}

func sendSocketMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-send
		log.Println("new message")
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func sendNatsMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-receive
		log.Println("message to nats")
		// Send it out to every client that is currently connected
		c.Publish("msg.frontend", msg)
		c.Flush()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	check(err)
	defer socket.Close()
	clients[socket] = true
	for {
		var msg frontendMsg
		err := socket.ReadJSON(&msg)
		fmt.Println(msg)
		if err != nil {
			delete(clients, socket)
			log.Panic(err.Error())
		}
		receive <- &msg
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
