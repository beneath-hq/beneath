package gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// ListenAndServeWS serves a WebSocket API
func ListenAndServeWS(port int) error {
	hub := newHub()
	go hub.run()

	router := mux.NewRouter()
	// router.HandleFunc("/", serveHome) // do we want something like this? see https://github.com/gorilla/websocket/edit/master/examples/chat/main.go for an example of the serveHome fn
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Printf("WS server running on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

// browsers will subscribe to websockets of: stream instanceid
// we will ultimately have multiple servers, so need to ensure that the websockets will scale
// question: do we have to worry about the clients sending the gateway server malicious data via the websocket?

// EXTRA CODE

// // establish websocket connection with each client
// func serveWs(w http.ResponseWriter, r *http.Request) {
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// register client
// 	clients[ws] = true

// 	// enable writes to each client
// 	go writeToClient()

// 	// enable reads from each client
// 	go readFromClient()
// }

/*
// emit data to client connections
func writeToClient() {
	for {
		time.Sleep(3000 * time.Millisecond)

		// generate fake data
		randNums := fmt.Sprint(rand.Intn(100), rand.Intn(100))

		// send data to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(randNums))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}

			log.Printf("finished sending data to one client")
		}
		log.Printf("finished sending data to all clients")
	}
}

func readFromClient() {
	for {
		time.Sleep(3000 * time.Millisecond)

		// listen to every client that is currently connected
		for client := range clients {
			log.Printf("waiting on a message...")
			_, p, err := client.ReadMessage()
			log.Print(string(p))
			if err != nil {
				log.Println(err)
			}
			log.Printf("finished listening to one client")
		}
		log.Printf("finished listening to all clients")
	}
}
*/

// each gateway needs its own subscription name to redis

// get reads working
// go routine for each client
// put code in place
// get data from redis... framework

/*
import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// func hello(w http.ResponseWriter, r *http.Request) {
// 	io.WriteString(w, "Hello GopherCon Israel 2019!")
// }

func ws(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to a websocket connection
	upgrader := websocket.Upgrader{} // ReadBufferSize: 1024; WriteBufferSize: 1024
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	for {
		// Read messages from socket
		// _, msg, err := conn.ReadMessage()
		// if err != nil {
		// 	log.Printf("Failed to read message %v", err)
		// 	conn.Close()
		// 	return
		// }
		// log.Printf("msg: %s", string(msg))

		// // write messages from socket
		// wait 7 seconds
		time.Sleep(7 * time.Second)
		messageType := websocket.TextMessage
		data := []byte{1, 2, 3}
		if err = conn.WriteMessage(messageType, data); err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", ws)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

// very helpful: https://github.com/eranyanay/1m-go-websockets
// scaling one server:
// problem 1. go kernel file limit is 1024 open files at once; each socket is represented by a file.
// solution 1: override the artificial file limit with a much larger number
// problem 2. at relatively small numbers of connections it will use up tons of RAM. in the example, 50k connections used up 1 GB of RAM.; each connection consumes ~20KB; 1M connections would take 20 GB of RAM
// solution 2: use Epoll (async I/O) to reduce the # of goroutines; reuse goroutines and reduce memory footprint; use epolls to do this; in the example, goroutines reduced from 50k to 5; now using 600 MB of RAM
// problem 3: buffers allocations; don't need a bufio connection because we are doing all input/output asynchronously; gorilla/websocket keeps a reference to the underlying buffers given by Hijack()
// solution 3: use more performant websockets library: github.com/gobwas/ws; this reduces buffer allocations; in the example, now using 60 MB of RAM for 50k connections.
// problem 4: conntrack table is full
// solution 4: change the server's settings to increase the cap of the total concurrent connections in the operating system
*/
/*
// solution 1 to override resource limitations:
// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
*/
