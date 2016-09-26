package operation

import (
	"log"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

// Server ...
type Server struct {
	Host     *string `json:"host,omitempty" bson:",omitempty"`
	User     *string `json:"user,omitempty" bson:",omitempty"`
	Group    *string `json:"group,omitempty" bson:",omitempty"`
	Password *string `json:"password,omitempty" bson:",omitempty"`
	Path     *string `json:"path,omitempty" bson:",omitempty"`
}

// Project ...
type Project struct {
	Name        *string     `json:"name,omitempty" bson:",omitempty"`
	Backend     *Repository `json:"backend,omitempty" bson:",omitempty"`
	Front       *Repository `json:"front,omitempty" bson:",omitempty"`
	Description *string     `json:"description,omitempty" bson:",omitempty"`
}

// Repository ...
type Repository struct {
	Address *string `json:"address"`
	Branch  *string `json:"branch"`
}

// Supervisor ...
type Supervisor struct {
	Path      *string `json:"path,omitempty" bson:",omitempty"`
	Include   *string `json:"include,omitempty" bson:",omitempty"`
	Extension *string `json:"extension,omitempty" bson:",omitempty"`
}

// Gunicorn ...
type Gunicorn struct {
	Workers     *string `json:"workers,omitempty" bson:",omitempty"`
	WorkerClass *string `json:"workerClass,omitempty" bson:"worker_class,omitempty"`
	Program     *string `json:"program,omitempty" bson:",omitempty"`
	Bind        *string `json:"bind,omitempty" bson:",omitempty"`
}

// Task ...
type Task struct {
	ID           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name         *string       `json:"name,omitempty" bson:",omitempty"`
	RemoteServer *Server       `json:"remoteServer,omitempty" bson:"remote_server,omitempty"`
	Project      *Project      `json:"project,omitempty" bson:",omitempty"`
	Supervisor   *Supervisor   `json:"supervisor,omitempty" bson:",omitempty"`
	Gunicorn     *Gunicorn     `json:"gunicorn,omitempty" bson:",omitempty"`
	VenvPath     *string       `json:"venvPath,omitempty" bson:"venv_path,omitempty"`
	LocalServer  *Server       `json:"localServer,omitempty" bson:"local_server,omitempty"`
	Nginx        *string       `json:"nginx,omitempty" bson:",omitempty"`
	ConfigName   *string       `json:"configName,omitempty" bson:"config_name,omitempty"`
}

func createTask(c *iris.Context) {
	newTask := Task{}
	err := c.ReadJSON(&newTask)
	if err != nil {
		log.Fatal(err)
	}
	coll := db.C("tasks")
	newTask.ID = bson.NewObjectId()
	err = coll.Insert(&newTask)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.JSON(201, &newTask); err != nil {
		log.Fatal(err)
	}
}

func queryTask(c *iris.Context) {
	coll := db.C("tasks")
	var tasks []Task
	task := Task{}
	iter := coll.Find(nil).Iter()
	for iter.Next(&task) {
		tasks = append(tasks, task)
	}
	if err := c.JSON(200, tasks); err != nil {
		log.Fatal(err)
	}
}

func updateTask(c *iris.Context) {
	coll := db.C("tasks")
	taskID := c.Param("taskID")
	updateValue := Task{}
	c.ReadJSON(&updateValue)
	if err := coll.UpdateId(bson.ObjectIdHex(taskID), bson.M{"$set": updateValue}); err != nil {
		log.Fatal(err)
	}
	err := coll.FindId(bson.ObjectIdHex(taskID)).One(&updateValue)
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(200, &updateValue)
}

func deleteTask(c *iris.Context) {
	coll := db.C("tasks")
	taskID := c.Param("taskID")
	if err := coll.RemoveId(bson.ObjectIdHex(taskID)); err != nil {
		log.Fatal(err)
	}
	c.SetStatusCode(204)
}
