package main

import "github.com/syfun/operation"

func main() {
	app := operation.CreateApp()
	app.Listen(":9000")
}
