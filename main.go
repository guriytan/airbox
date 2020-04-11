package main

import (
	. "airbox/config"
	"fmt"
)

func main() {
	Init()
	router := NewRouter().Init().PathMapping()
	err := router.Start(Env.Web.Port)
	if err != nil {
		fmt.Println(err)
	}
}
