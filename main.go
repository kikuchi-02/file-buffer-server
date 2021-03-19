package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kikuchi-02/file-buffer-server/libs"
)

var DB_SETTINGS libs.DBSettings

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	source := libs.BufferSetup()
	settings := libs.LoadSettings("settings.yaml")
	libs.LoadDBSettings("user.yaml")

	http.HandleFunc(settings.Endpoint, libs.EventlogHander(source))

	log.Printf("serve on Port:%d\n", settings.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil)

}
