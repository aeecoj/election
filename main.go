package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	_ "github.com/go-sql-driver/mysql"
)

type client struct{} // Add more data to this type if needed

var clients = make(map[*websocket.Conn]client) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)

func runHub() {

	for {
		select {
		case connection := <-register:
			clients[connection] = client{}
			log.Println("connection registered")

		case message := <-broadcast:
			log.Println("message received:", message)

			// Send the message to all clients
			for connection := range clients {
				if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Println("write error:", err)

					unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(clients, connection)

			log.Println("connection unregistered")
		}
	}
}

func main() {
	app := fiber.New()

	app.Static("/", "./home.html")

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	go runHub()

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {

		// When the function returns, unregister the client and close the connection
		defer func() {
			unregister <- c
			c.Close()
		}()

		// Register the client
		register <- c

		for {

			//var result []DikaCourt

			// messageType, _, err := c.ReadMessage()
			// if err != nil {
			// 	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// 		log.Println("read error:", err)
			// 	}

			// 	return // Calls the deferred function, i.e. closes the connection on error
			// }
			// log.Println(count)
			// if messageType == websocket.TextMessage {
			// 	// Broadcast the received message
			// 	// broadcast <- string(string(message))
			// 	// broadcast <- string(string(out))
			// 	broadcast <- string(string(count))
			// } else {
			// 	log.Println("websocket message received of type", messageType)
			// }

			//count := 0
			for range time.Tick(time.Second * 10) {
				//count++
				result := Supreme()
				resultByte, err := json.Marshal(result)
				if err != nil {
					panic(err)
				}
				broadcast <- string(resultByte)

			}

		}
	}))

	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()
	log.Fatal(app.Listen(*addr))
}

type DikaCourt struct {
	DNum         string `json:"d_num"`
	Pname        string `json:"p_num"`
	Fname        string `json:"f_num"`
	Lname        string `json:"l_num"`
	Votes        int    `json:"votes"`
	ResponseTime string `json:"response_time"`
}

func Supreme() []DikaCourt {
	fmt.Println("Go MySQL Tutorial")

	// Open up our database connection.
	// I've set up a database on my local machine using phpmyadmin.
	// The database is called testDb
	db, err := sql.Open("mysql", "root:xibPPk27@tcp(10.1.2.28:3306)/votes")

	// if there is an error opening the connection, handle it
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	// Execute the query
	results, err := db.Query("SELECT D_num,Pname,Fname, Lname, votes FROM dika_court ORDER BY votes DESC LIMIT 2")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	var dikaCourtList []DikaCourt
	for results.Next() {
		var dikaCourt DikaCourt
		// for each row, scan the result into our tag composite object
		err = results.Scan(&dikaCourt.DNum, &dikaCourt.Pname, &dikaCourt.Fname, &dikaCourt.Lname, &dikaCourt.Votes)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		//log.Printf(dikaCourt.DNum, dikaCourt.Pname, dikaCourt.Fname, dikaCourt.Lname, dikaCourt.Votes)

		dikaCourt.ResponseTime = time.Now().String()
		dikaCourtList = append(dikaCourtList, dikaCourt)
	}

	return dikaCourtList
}
