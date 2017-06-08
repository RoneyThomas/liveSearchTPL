package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"encoding/csv"
	"encoding/json"
	"strconv"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "45.55.209.67:4571", "socket ip")

type Json struct {
	Terms string `json:"terms"`
	Browser string `json:"browser"`
}

func listenScoket() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/rtsearches"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	checkError("dial:", err)
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			checkError("read:", err)
			//json to csv
			d := Json{}
			x := message[1:len(message)-1]
			log.Printf("recv: %s", message)
			err = json.Unmarshal(x, &d)
			checkError("json unmarshal:", err)
			//File write
			writeToCSV(d)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			checkError("Write", err)
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			checkError("Write close", err)
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

// Func to write in CSV format
func writeToCSV(message Json) {
	file, err := os.OpenFile("searchHistory.csv", os.O_APPEND|os.O_WRONLY, 0600)
	checkError("Cannot create file", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Get Unix epoch time in nanosecond
	nanos := strconv.FormatInt(time.Now().UnixNano(), 10)
	var record []string
	record = append(record, nanos)
	record = append(record, message.Terms)
	record = append(record, message.Browser)
	// Write record in CSB format
	writer.Write(record)
}

// Func to check error and logs them
func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
