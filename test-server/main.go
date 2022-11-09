package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":9090", "http listen address")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hBytes, _ := json.Marshal(r.Header)
		body := r.Body
		if body != nil {
			defer body.Close()
		}

		bodyBytes, _ := ioutil.ReadAll(body)

		w.Write([]byte("welcome to shop! \n"))
		fmt.Println("header:" + string(hBytes) + "\n")
		w.Write([]byte("header:" + string(hBytes) + "\n"))
		fmt.Println("body:" + string(bodyBytes) + "\n")
		w.Write([]byte("body:" + string(bodyBytes) + "\n"))

	})

	log.Printf("start success! listen address is %+v", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
