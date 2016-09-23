package operation

import (
	"log"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2"
)

var db *mgo.Database

func initDB(dbURL string) {
	session, err := mgo.Dial(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB("operation")
}

// CreateApp ...
func CreateApp() *iris.Framework {
	initDB("localhost")
	app := iris.New()
	app.Get("/", func(c *iris.Context) {
		c.WriteString("hello operation")
	})
	app.Post("/servers", createServer)
	app.Get("/servers", queryServer)
	app.Post("/servers/:serverID", updateServer)
	return app
}
