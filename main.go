package main

import (
	"context"
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

func createServer(dir string) http.Server {
	mux := http.NewServeMux()
	fserver := http.FileServer(http.Dir(dir))
	mux.Handle("/", prettyPrint(fserver))
	return http.Server{
		Addr:              "localhost:8080",
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
		err       error
	)

	if len(os.Args) > 1 {
		directory, err = filepath.Abs(os.Args[1])
		if err != nil {
			log.Println(err)
			directory, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Using current working directory instead.")
		}
	} else {
		directory, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	allConnsClosed := make(chan struct{})
	srv := createServer(directory)
	go shutdownServerOnSignal(&srv, allConnsClosed)

	fmt.Printf("Serving files from \033[0;33m%s\033[0m\n", directory)

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("HTTP Server Error: ", err)
	}

	<-allConnsClosed
	fmt.Println("Successfully shutdown. Bye.")
}
