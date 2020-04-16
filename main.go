package main

import (
	. "airbox/config"
)

func main() {
	Init()
	router := NewRouter().Init().PathMapping()
	router.Logger.Fatal(router.Start(Env.Web.Port))
}
