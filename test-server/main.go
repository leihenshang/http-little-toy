package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":9090", "http listen address")
	useHttps   = flag.Bool("useHttps", false, "use https")
	useSsl     = flag.Bool("useHttps", false, "use https")
)

func main() {
	flag.Parse()

	if *useSsl {
		cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
		if err != nil {
			log.Println(err)
			return
		}

		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		ln, err := tls.Listen("tcp", ":443", config)
		if err != nil {
			log.Println(err)
			return
		}
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go handleConn(conn)
		}
	}

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

func handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		println(msg)
		n, err := conn.Write([]byte("world\n"))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
