package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mudkipme/lilycove/lib"
	"github.com/mudkipme/lilycove/routers"
)

var serveMux *http.ServeMux

func main() {
	serveMux = http.NewServeMux()
	routers.Init(serveMux)
	config := lib.Config()
	port := config.HTTP.Port
	if port == 0 {
		port = 8080
	}
	server := &http.Server{Addr: ":" + strconv.Itoa(config.HTTP.Port), Handler: serveMux}
	log.Fatal(server.ListenAndServe())
}
