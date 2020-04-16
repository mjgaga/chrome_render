package wsserver

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	fd,_:=os.OpenFile(fmt.Sprintf("%d.webm", time.Now().Unix()),os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			fd.Close()
			break
		}

		log.Printf("recv type: %d, data len: %d", mt,  len( message))
		fd.Write(message)
	}
}

func Start() {
	http.HandleFunc("/", echo)
	logrus.Fatal(http.ListenAndServe("localhost:8080", nil))
}