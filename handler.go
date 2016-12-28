package operation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
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
	ID          *int        `json:"id"`
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
	Tags         []string      `json:"tags"`
	Group        bson.ObjectId `json:"group_id" bson:"group_id"`
}

// Group for tasks.
type Group struct {
	ID         bson.ObjectId   `json:"id" bson:"_id,omitempty"`
	Name       string          `json:"name"`
	Tasks      []bson.ObjectId `json:"tasks"`
	FrontTag   string          `json:"front_tag"`
	BackendTag string          `json:"backend_tag"`
	CMSTag     string          `json:"cms_tag"`
}

func createTask(c *iris.Context) {
	session := gSession.Clone()
	defer session.Close()
	newTask := Task{}
	err := c.ReadJSON(&newTask)
	if err != nil {
		log.Fatal(err)
	}
	coll := session.DB("operation").C("tasks")
	newTask.ID = bson.NewObjectId()
	err = coll.Insert(&newTask)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.JSON(201, &newTask); err != nil {
		log.Fatal(err)
	}
}

// Tag ...
type Tag struct {
	Name string `json:"name"`
}

func getGroups(ctx *iris.Context)  {
	session := gSession.Clone()
	defer session.Close()
	coll := session.DB("operation").C("groups")
	groups := make([]Group, 3)
	err := coll.Find(nil).All(&groups)
	if err != nil {
		log.Fatal(err)
	}
	ctx.JSON(200, groups)
}

func getTags(projectName string) ([]Tag, error) {
	privateToken := viper.GetString("privateToken")
	url := fmt.Sprintf("http://dev.titangroupco.com/api/v3/projects?private_token=%v&search=%v", privateToken, projectName)
	statusCode, body, err := fasthttp.Get(nil, url)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.New(string(body))
	}
	var projects []Project
	err = json.Unmarshal(body, &projects)
	if err != nil {
		return nil, err
	}
	if len(projects) < 1 {
		return nil, fmt.Errorf("Cannot find %v project.", projectName)
	}

	url = fmt.Sprintf("http://dev.titangroupco.com/api/v3/projects/%v/repository/tags?private_token=%v", *projects[0].ID, privateToken)
	statusCode, body, err = fasthttp.Get(nil, url)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.New(string(body))
	}
	var tags []Tag
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func queryTask(c *iris.Context) {
	session := gSession.Clone()
	defer session.Close()
	coll := session.DB("operation").C("tasks")
	var tasks []Task
	task := Task{}

	groupID := c.URLParam("group")

	query := make(bson.M)
	if groupID != "" {
		query = bson.M{"group_id": bson.ObjectIdHex(groupID)}
	} else {
		query = nil
	}

	iter := coll.Find(query).Iter()
	proTagsMap := make(map[string][]string, 0)
	for iter.Next(&task) {
		projectName := *task.Project.Name
		proTags, ok := proTagsMap[projectName]
		if !ok {
			tags, err := getTags(projectName)
			if err != nil {
				log.Println(err)
			}
			proTags = make([]string, 0)
			for _, tag := range tags {
				proTags = append(proTags, tag.Name)
			}
		}
		task.Tags = proTags
		tasks = append(tasks, task)
	}
	if err := iter.Close(); err != nil {
		log.Panic(err)
	}
	if err := c.JSON(200, tasks); err != nil {
		log.Panic(err)
	}
}

func updateTask(c *iris.Context) {
	session := gSession.Clone()
	defer session.Close()
	coll := session.DB("operation").C("tasks")
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
	session := gSession.Clone()
	defer session.Close()
	coll := session.DB("operation").C("tasks")
	taskID := c.Param("taskID")
	if err := coll.RemoveId(bson.ObjectIdHex(taskID)); err != nil {
		log.Fatal(err)
	}
	c.SetStatusCode(204)
}
