package operation

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	json "github.com/bitly/go-simplejson"
	"github.com/kataras/iris"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var db *mgo.Database

func initDB(dbURL string) {
	session, err := mgo.Dial(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB("operation")
}

func getArgs(task *Task) []string {
	fabPath := "/home/sunyu/gowork/src/github.com/syfun/operation/fabfile.py"
	deploy := fmt.Sprintf("deploy:tmp_path=%s,backend_url=%s,backend_branch=%s,front_url=%s,front_branch=%s,remote_path=%s,venv_path=%s,program=%s,workers=%s,worker_class=%s,bind=%s,user_group=%s,ext=%s,path=%s,include=%s,local_user=%s,local_password=%s,config_name=%s,nginx=%v",
		*task.LocalServer.Path, *task.Project.Backend.Address, *task.Project.Backend.Branch, *task.Project.Front.Address, *task.Project.Front.Branch, *task.RemoteServer.Path, *task.VenvPath, *task.Gunicorn.Program, *task.Gunicorn.Workers,
		*task.Gunicorn.WorkerClass, *task.Gunicorn.Bind, *task.RemoteServer.Group, *task.Supervisor.Extension, *task.Supervisor.Path, *task.Supervisor.Include, *task.LocalServer.User, *task.LocalServer.Password, *task.ConfigName, *task.Nginx)
	cmd := []string{
		"-f", fabPath, "-u", *task.RemoteServer.User, "-p", *task.RemoteServer.Password, "-H", *task.RemoteServer.Host, deploy}
	return cmd
}

// RunCommand ...
func RunCommand(c iris.WebsocketConnection, taskID string) *exec.Cmd {
	task := &Task{}
	coll := db.C("tasks")
	err := coll.FindId(bson.ObjectIdHex(taskID)).One(task)
	if err != nil {
		log.Fatal(err)
	}
	cmdArgs := getArgs(task)
	cmd := exec.Command("fab", cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			c.EmitMessage(scanner.Bytes())
		}
	}()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return cmd
}

// CreateApp ...
func CreateApp() *iris.Framework {
	initDB("192.168.0.238")
	app := iris.New()
	app.Get("/", func(c *iris.Context) {
		c.WriteString("hello operation")
	})
	app.Post("/api/v1/tasks", createTask)
	app.Get("/api/v1/tasks", queryTask)
	app.Put("/api/v1/tasks/:taskID", updateTask)
	app.Delete("/api/v1/tasks/:taskID", deleteTask)

	app.Config.Websocket.Endpoint = "/ws"
	app.Websocket.OnConnection(func(c iris.WebsocketConnection) {
		fmt.Println("Connected.")
		var cmd *exec.Cmd
		c.OnMessage(func(message []byte) {
			fmt.Println(string(message))
			js, _ := json.NewJson(message)
			msgType, err := js.Get("type").String()
			if err != nil {
				log.Fatal(err)
			}
			if msgType == "deploy" {
				taskID, err := js.Get("taskID").String()
				if err != nil {
					log.Fatal(err)
				}
				cmd = RunCommand(c, taskID)
			} else if msgType == "stop" {
				if err := cmd.Process.Kill(); err != nil {
					log.Fatal(err)
				}
			}
		})
	})
	return app
}
