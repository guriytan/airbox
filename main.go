package main

import (
	"airbox/global"
	"log"
)

func main() {
	router := NewRouter().PathMapping()
	log.Fatal(router.Start(global.Env.Web.Port))
}
