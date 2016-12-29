package operation

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)


//func TestExecCmd(t *testing.T) {
//	fmt.Println("Start test TestExecCmd...")
//	cmd := exec.Command("sleep", "5")
//	err := cmd.Start()
//	if err != nil {
//		t.Error(err)
//	}
//	log.Println("Waiting for command to finish...")
//	err = cmd.Wait()
//	log.Printf("Command finished with error: %v", err)
//}

//func TestUpdateTag(t *testing.T) {
//	fmt.Println("Start test TestUpdateTag...")
//	session := gSession.Clone()
//	defer session.Close()
//	var task Task
//	coll := session.DB("operation").C("tasks")
//
//	// kelvin
//	if err := coll.Find(bson.M{"project.name": "kelvin"}).One(&task); err != nil {
//		t.Error(err)
//	}
//	updateTag(&task, "front", "back")
//
//	// cms_plm
//	if err := coll.Find(bson.M{"project.name": "cms_plm"}).One(&task); err != nil {
//		t.Error(err)
//	}
//	updateTag(&task, "", "cms_test")
//}

func TestTmpDir(t *testing.T)  {
	dir, err := ioutil.TempDir("D:\\", "operation")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(dir)
}

func TestMain(m *testing.M) {
	var err error
	gSession, err = initDB("192.168.0.238:27017")
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}