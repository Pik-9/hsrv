// Copyright (c) 2022 Daniel Steinhauer
//
// Licensed under the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func prettyPrint(next http.Handler) http.Handler {
	color := map[string]int{
		"GET":    32,
		"POST":   33,
		"PUT":    35,
		"PATCH":  34,
		"DELETE": 31,
		"HEAD":   36,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cl, ok := color[r.Method]
		if !ok {
			cl = 30
		}

		whitespaces := strings.Repeat(" ", 8-len(r.Method))

		fmt.Printf("\033[1;%dm%s%s\033[0m%s\n", cl, r.Method, whitespaces, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func createServer(dir string, port int) http.Server {
	mux := http.NewServeMux()
	fserver := http.FileServer(http.Dir(dir))
	mux.Handle("/", prettyPrint(fserver))
	return http.Server{
		Addr:              fmt.Sprintf("localhost:%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       3 * time.Second,
		IdleTimeout:       3 * time.Second,
	}
}

func shutdownServerOnSignal(srv *http.Server, connsClosed chan<- struct{}) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	signum := <-sigchan

	fmt.Printf("Received signal #%d. Shutting down...\n", signum)

	srvErr := srv.Shutdown(context.Background())
	if srvErr != nil {
		log.Fatal(srvErr)
	}

	close(connsClosed)
}

func main() {
	var (
		directory string
		port      int
		err       error
	)

	curDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flag.IntVar(&port, "p", 8080, "The port to listen on.")
	flag.Parse()

	if flag.Arg(0) == "" {
		directory = curDir
	} else {
		directory, err = filepath.Abs(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	}

	allConnsClosed := make(chan struct{})
	srv := createServer(directory, port)
	go shutdownServerOnSignal(&srv, allConnsClosed)

	fmt.Printf("Serving files from \033[0;33m%s\033[0m on \033[0;32m%s\033[0m\n", directory, srv.Addr)

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("HTTP Server Error: ", err)
	}

	<-allConnsClosed
	fmt.Println("Successfully shutdown. Bye.")
}
