package main

import (
	"airbox/config"
	"log"
)

func main() {
	router := NewRouter().PathMapping()
	log.Fatal(router.Start(config.Env.Web.Port))
}
