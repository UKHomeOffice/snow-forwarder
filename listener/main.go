package main

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
)

func main() {

	http.HandleFunc("/", Handler)
	log.Fatal(gateway.ListenAndServe("", nil))

}
