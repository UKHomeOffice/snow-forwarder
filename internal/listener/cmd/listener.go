package main

import (
	"log"

	"github.com/UKHomeOffice/snow-forwarder/internal/listener"
	"github.com/apex/gateway"
)

func main() {

	log.Fatal(gateway.ListenAndServe("", listener.Handler()))
}
