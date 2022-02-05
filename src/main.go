// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package signalgo

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

var addr = flag.String("addr", ":9099", "http service address")
var mimes = map[string]string{
	"js":   "application/javascript",
	"css":  "text/css",
	"html": "text/html",
}

func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	file := "./static"
	path := r.URL.Path
	if path != "/" {
		file += path
		temp := strings.Split(path, ".")
		ext := strings.ToLower(temp[len(temp)-1])
		w.Header().Set("Content-Type", mimes[ext])
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, file)
}

func main() {
	flag.Parse()
	signalGo := NewSignalGo()
	redisBackPlane := NewRedisBackplane("localhost:6379", "", 5, false)
	redisBackPlane.Init(signalGo)
	signalGo.UseBackplane(&redisBackPlane)
	go signalGo.Run()
	http.HandleFunc("/", serveStaticFiles)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(signalGo, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
