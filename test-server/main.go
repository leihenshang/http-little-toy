package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":9090", "http listen address")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome to shop! \n"))
	})

	log.Printf("start success! listen address is %+v", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
