package operation

import (
	"fmt"
	"log"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

// Server ...
type Server struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Host     string        `json:"host"`
	User     string        `json:"user"`
	Group    string        `json:"group"`
	Password string        `json:"password"`
}

func createServer(c *iris.Context) {
	newServer := Server{}
	err := c.ReadJSON(&newServer)
	if err != nil {
		log.Fatal(err)
	}
	coll := db.C("servers")
	newServer.ID = bson.NewObjectId()
	err = coll.Insert(&newServer)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.JSON(201, &newServer); err != nil {
		log.Fatal(err)
	}
}

func queryServer(c *iris.Context) {
	coll := db.C("servers")
	var servers []Server
	server := Server{}
	iter := coll.Find(nil).Iter()
	for iter.Next(&server) {
		servers = append(servers, server)
	}
	if err := c.JSON(200, servers); err != nil {
		log.Fatal(err)
	}
}

func updateServer(c *iris.Context) {
	coll := db.C("servers")
	serverID := c.Param("serverID")
	updateValue := Server{}
	c.ReadJSON(&updateValue)
	fmt.Println(serverID)
	err := coll.UpdateId(bson.ObjectIdHex(serverID), bson.M{"$set": updateValue})
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(200, &updateValue)
}
