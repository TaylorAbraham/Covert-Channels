package main

import (
	"./controller"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"strconv"
)

func main() {

	//Intercept the kill signal to ensure proper shutdown
	//of the process
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)

  var p *int = flag.Int("p", 6000, "the port for the websocket ")
  flag.Parse()

  ctr, err := controller.CreateController()
  if err != nil {
    log.Fatal(err.Error())
  }
	//Create each of the possible websocket connections
	mux := http.NewServeMux()

	mux.HandleFunc("/", ctr.HandleFunc)

	defer ctr.Shutdown()

	srv := &http.Server{Addr: ":" + strconv.Itoa(*p), Handler: mux}

	log.Println("http server started on :" + strconv.Itoa(*p))

	//Go routine to listen to kill signals for this process
	go func() {
		<-signalChan
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(ctx)
	}()

	//Start and listen for websocket connections
	err = srv.ListenAndServe()
	if err != nil {
		log.Println("ListenAndServer: ", err)
	}
	log.Println("Shutting down")
}
