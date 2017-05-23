package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	nats "github.com/nats-io/go-nats"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type frontendMsg struct {
	Msg_type     int `json:"Msg_type"`
	ID           int `json:"ID"`
	Lat          int `json:"Lat"`
	Lng          int `json:"Lng"`
	CauseCode    int `json:"CauseCode"`
	SubCauseCode int `json:"SubCauseCode"`
	Timestamp    int `json:"Timestamp"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan *frontendMsg)

func main() {
	nc, err := nats.Connect("nats://193.136.93.80:4222")
	check(err)
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	check(err)
	defer c.Close()
	c.Subscribe("msg.frontend", func(p *frontendMsg) {
		fmt.Println(p)
		broadcast <- p
	})

	http.HandleFunc("/", handler)
	go handleMessages()

	log.Println("http server started on :4000")
	err = http.ListenAndServe(":4000", nil)
	check(err)

}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
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
		broadcast <- &msg
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
func main() {
	session, err := nats.Connect("nats;//localhost:4222")

	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)

	router.Handle("denm add", addChannel)
	router.Handle("denm remove", subscribeChannel)

	router.Handle("cam add", unsubscribeChannel)
	router.Handle("cam remove", editUser)

	router.Handle("message add", addChannelMessage)
	router.Handle("message subscribe", subscribeChannelMessage)
	router.Handle("message unsubscribe", unsubscribeChannelMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
*/
// type Channel struct {
//   Id string `json:"id" gorethink:"id,omitempty"`
//   Name string `json:"name" gorethink:"name"`
// }
//
// type User struct {
//   Id string `gorethink:"id,omitempty"`
//   Name string `gorethink:"name"`
// }

// func handler(w http.ResponseWriter, r *http.Request) {
//   // fmt.Fprintf(w, "hello from go")
//   // var socket *websocket.Conn
//   // var err error
//   // var socket, err = upgrader.Upgrade(w, r, nil)
//
//   if err != nil {
//     fmt.Println(err)
//     return
//   }
//   for {
//     // msgType, msg, err := socket.ReadMessage()
//     // if err != nil {
//     //   fmt.Println(err)
//     //   return
//     // }
//     var inMessage Message
//     var outMessage Message
//     if err := socket.ReadJSON(&inMessage); err != nil {
//       fmt.Println(err)
//       break
//     }
//     fmt.Printf("%#v\n", inMessage)
//     switch inMessage.Name {
//     case "channel add":
//       err := addChannel(inMessage.Data)
//       if err != nil {
//         outMessage = Message{"error", err}
//         if err := socket.WriteJSON(outMessage); err != nil {
//           fmt.Println(err)
//           break
//         }
//       }
//     case "channel subscribe":
//       go subscribeChannel(socket)
//     }
//     // fmt.Println(string(msg))
//     // if err = socket.WriteMessage(msgType, msg); err != nil {
//     //   fmt.Println(err)
//     //   return
//     // }
//   }
// }
//do not communicate by sharing memory, share memory by communicating
// func addChannel(data interface{}) (error) {
//   var channel Channel
//
//   err := mapstructure.Decode(data, &channel)
//   if err != nil {
//     return err
//   }
//   channel.Id = "1"
//   // fmt.Printf("%#v\n", channel)
//   fmt.Println("add channel")
//   return nil
// }
//
// func subscribeChannel(socket *websocket.Conn) {
//   //simulate db query / changefeed
//   for {
//     time.Sleep(time.Second * 1)
//     message := Message{"channel add",
//       Channel{"1", "software support"}}
//     socket.WriteJSON(message)
//     fmt.Println("sent new channel")
//   }
// }
