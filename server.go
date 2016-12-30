package operation

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	json "github.com/bitly/go-simplejson"
	"github.com/iris-contrib/middleware/cors"
	"github.com/iris-contrib/middleware/logger"
	"github.com/iris-contrib/middleware/recovery"
	"github.com/kataras/go-template/html"
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var gSession *mgo.Session

func initDB(dbURL string) (*mgo.Session, error) {
	session, err := mgo.Dial(dbURL)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to mongo, %v", err)
	}
	return session, err
}

func getArgs(task *Task, frontTag, backTag string) []string {
	fabPath := viper.GetString("fabPath")
	if frontTag == "" {
		frontTag = task.Project.Front.Branch
	}
	if backTag == "" {
		backTag = task.Project.Backend.Branch
	}
	tmp, err := ioutil.TempDir("/tmp/", "operation")
	if err != nil {
		log.Panic(err)
	}
	deploy := fmt.Sprintf("deploy:tmp_path=%s,backend_url=%s,backend_branch=%s,front_url=%s,front_branch=%s,remote_path=%s,venv_path=%s,program=%s,workers=%s,worker_class=%s,bind=%s,user_group=%s,ext=%s,path=%s,include=%s,local_user=%s,local_password=%s,config_name=%s,nginx=%v",
		tmp, task.Project.Backend.Address, backTag, task.Project.Front.Address, frontTag, task.RemoteServer.Path, task.VenvPath, task.Gunicorn.Program, task.Gunicorn.Workers,
		task.Gunicorn.WorkerClass, task.Gunicorn.Bind, task.RemoteServer.Group, task.Supervisor.Extension, task.Supervisor.Path, task.Supervisor.Include, task.LocalServer.User, task.LocalServer.Password, task.ConfigName, task.Nginx)
	cmd := []string{
		"-f", fabPath, "-u", task.RemoteServer.User, "-p", task.RemoteServer.Password, "-H", task.RemoteServer.Host, deploy}
	return cmd
}

func updateTag(t *Task, frontTag, backTag string) {
	session := gSession.Clone()
	defer session.Close()
	var group Group
	coll := session.DB("operation").C("groups")
	err := coll.FindId(t.Group).One(&group)
	if err != nil {
		log.Fatal(err)
	}
	switch t.Project.Name {
	case "kelvin":
		if err := coll.UpdateId(group.ID, bson.M{"$set": bson.M{"front_tag": frontTag, "back_tag": backTag}}); err != nil {
			log.Fatal(err)
		}

	case "cms_plm":
		if err := coll.UpdateId(group.ID, bson.M{"$set": bson.M{"cms_tag": backTag}}); err != nil {
			log.Fatal(err)
		}
	}
}

// RunCommand ...
func RunCommand(c iris.WebsocketConnection, taskID, frontTag, backTag string) {
	session := gSession.Clone()
	defer session.Close()
	var task Task
	coll := session.DB("operation").C("tasks")
	err := coll.FindId(bson.ObjectIdHex(taskID)).One(&task)
	if err != nil {
		log.Fatal(err)
	}

	cmdArgs := getArgs(&task, frontTag, backTag)
	cmd := exec.Command("fab", cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout
	scanner := bufio.NewScanner(stdout)
	ch := make(chan int)
	go func() {
		for scanner.Scan() {
			c.EmitMessage(scanner.Bytes())
		}
		ch <- 0
	}()
	if err := cmd.Start(); err != nil {
		log.Panic(fmt.Errorf("Exec error, %v", err))
	}
	if err := cmd.Wait(); err != nil {
		log.Panic(fmt.Errorf("Finish error, %v", err))
	}
	log.Println("Deploy Over.")
	updateTag(&task, frontTag, backTag)
	<-ch
	c.EmitMessage([]byte("Deploy Over."))
}

// CreateApp ...
func CreateApp() *iris.Framework {
	var err error
	viper.AddConfigPath("/Users/sunyu/workspace/goprojects/src/github.com/syfun/operation")
	viper.AddConfigPath("D:/Workspace/gowork/src/github.com/syfun/operation")
	viper.AddConfigPath("/opt/operation")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	if err = viper.ReadInConfig(); err != nil {
		log.Panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	gSession, err = initDB(viper.GetString("mongoURL"))
	if err != nil {
		log.Fatal(err)
	}

	app := iris.New()
	fmt.Println("##########")
	fmt.Println("templtePath", viper.GetString("templatePath"))
	fmt.Println("fabPath", viper.GetString("fabPath"))
	app.UseTemplate(html.New()).Directory(viper.GetString("templatePath"), ".html")
	app.Static("/static", viper.GetString("staticPath"), 1)

	app.Use(recovery.Handler)

	crs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE"},
	})
	app.Use(crs)

	errorLogger := logger.New(logger.Config{Status: true, IP: true, Method: true, Path: true})
	app.Use(errorLogger)

	app.Get("/", func(c *iris.Context) {
		c.MustRender("index.html", nil)
	})
	app.Post("/api/v1/tasks", createTask)
	app.Get("/api/v1/tasks", queryTask)
	//app.Put("/api/v1/tasks/:taskID", updateTask)
	//app.Delete("/api/v1/tasks/:taskID", deleteTask)
	app.Get("/api/v1/groups", getGroups)

	app.Config.Websocket.Endpoint = "/ws"
	app.Websocket.OnConnection(func(c iris.WebsocketConnection) {
		fmt.Println("Connected.")
		c.OnMessage(func(message []byte) {
			fmt.Println(string(message))
			js, _ := json.NewJson(message)
			msgType, err := js.Get("type").String()
			if err != nil {
				log.Panic(err)
			}
			if msgType == "deploy" {
				taskID, err := js.Get("taskID").String()
				if err != nil {
					log.Println(err)
				}
				frontTag, err := js.Get("frontTag").String()
				if err != nil {
					log.Println(err)
				}
				backTag, err := js.Get("backTag").String()
				if err != nil {
					log.Println(err)
				}
				RunCommand(c, taskID, frontTag, backTag)
			}
		})
	})
	return app
}
