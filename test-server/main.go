package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":9090", "http listen address")
	certFile   = flag.String("certFile", "", "cert file")
	keyFile    = flag.String("keyFile", "", "cert key")
	http2      = flag.Bool("h2", false, "use http2 protocol")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hBytes, err := json.Marshal(r.Header)
		if err != nil {
			log.Printf("marshal header error: %v", err)
			http.Error(w, "Failed to marshal header", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("read body error: %v", err)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("welcome to toy testing! \n"))
		w.Write([]byte("header:" + string(hBytes) + "\n"))
		w.Write([]byte("body:" + string(bodyBytes) + "\n"))

	})

	log.Printf("start test server on: %+v", *listenAddr)
	if *http2 {
		log.Fatal(http.ListenAndServeTLS(*listenAddr, *certFile, *keyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(*listenAddr, nil))
	}

}
