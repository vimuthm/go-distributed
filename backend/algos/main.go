package algos

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Node struct {
	id     int
	state  string
	round  int
	leader int
}

type Message struct {
	sender_id   int
	receiver_id int
	round       int
	message     int
}

func Start(ws *websocket.Conn) {
	chang_roberts(ws, 10)
}

func chang_roberts(ws *websocket.Conn, n int) {
	fmt.Println("Chang and Roberts")
	fmt.Println("n =", n)

	channels := make([]chan Message, n)
	nodes := make([]Node, n)
	done := make(chan bool)
	master_chans := make([]chan string, n)

	for i := 0; i < n; i++ {
		nodes[i].id = i
		nodes[i].state = "active"
		nodes[i].round = 0
		nodes[i].leader = -1
	}

	for i := 0; i < n; i++ {
		var left int
		if i == 0 {
			left = n - 1
		} else {
			left = i - 1
		}

		var right int
		if i == n-1 {
			right = 0
		} else {
			right = i
		}

		go node(i, channels[left], channels[right], done, left, right, master_chans[i])
	}

	// Listen to all master channels, on any receive write it to the websocket
	go func() {
		for {
			for _, master_chan := range master_chans {
				select {
				case msg := <-master_chan:
					fmt.Println("Received message on master channel")
					err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
					if err != nil {
						fmt.Println("Error writing to websocket")
					}
				default:
					// fmt.Println("No message received on master channel")
				}
			}
		}
	}()
}

// Message passing clockwise => left neighbour is sent to
func node(id int, write chan<- Message, read <-chan Message, done chan bool, left int, right int, master_chan chan<- string) {
	fmt.Println("Node ", id, " started")
	var msg = Message{
		sender_id:   id,
		receiver_id: left,
		round:       0,
		message:     id,
	}
	write <- msg
	fmt.Println("Node ", id, " sent message to ", left)

	for {
		select {
		case msg = <-read:
			// print := "Node ", id, " received message from ", msg.sender_id
			print := fmt.Sprintf("Node %d received message from %d", id, msg.sender_id)
			fmt.Println(print)
			master_chan <- print
			if msg.message == id {
				fmt.Println("Node ", id, " is the leader")
				done <- true
				return
			} else if msg.message > id {
				msg.round++
				write <- msg
			} else {
				print := fmt.Sprintf("Node %d dropped message %d from %d %d", id, msg.message, msg.sender_id, msg.round)
				fmt.Println(print)
				master_chan <- print
			}
		case <-done:
			return
		}
	}
}