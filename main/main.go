package main

import (
	"fmt"
	"log"

	"operation"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Book ...
type Book struct {
	Name    string   `json:"name"`
	Authors []string `json:"authors"`
}

func getDB() *Book {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("demo").C("books")
	err = c.Insert(&Book{"Cook", []string{"sdgasg", "sdgsdg"}},
		&Book{"Cla", []string{"ggg", "gggg"}})
	if err != nil {
		log.Fatal(err)
	}

	result := Book{}
	err = c.Find(bson.M{"name": "Cook"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Phone:", result.Authors)
	return &result
}

var db *mgo.Database

func main() {
	// session, err := mgo.Dial("localhost")
	// if err != nil {
	// 	panic(err)
	// }
	// db = session.DB("operation")
	// defer session.Close()

	// iris.Get("/", func(ctx *iris.Context) {
	// 	book := getDB()
	// 	ctx.JSON(iris.StatusOK, book)
	// })
	// iris.Post("/servers", createServer)

	// iris.Config.Websocket.Endpoint = "/ws"
	// iris.Websocket.OnConnection(func(c iris.WebsocketConnection) {
	// 	fmt.Println("Connected.")
	// 	c.OnMessage(func(message []byte) {
	// 		fmt.Println(string(message))
	// 		js, _ := json.NewJson(message)
	// 		fmt.Print(js.Get("type").String())
	// 		cmd := exec.Command("ping", "www.baidu.com")
	// 		stdout, err := cmd.StdoutPipe()
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		scanner := bufio.NewScanner(stdout)
	// 		go func() {
	// 			for scanner.Scan() {
	// 				c.EmitMessage(scanner.Bytes())
	// 			}
	// 		}()
	// 		if err := cmd.Start(); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		if err := cmd.Wait(); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	})
	// })

	// iris.Listen(":9000")
	app := operation.CreateApp()
	app.Listen(":9000")
}
