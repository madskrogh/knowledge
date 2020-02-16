package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	myhttp "raffle/knowledge/pkg/http"
	"raffle/knowledge/pkg/mongo"
)

func main() {
	f, err := os.OpenFile("app.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)
	ds, err := mongo.NewDocumentService("mongodb://db:27017", "knowledge", "one", l)
	if err != nil {
		log.Fatal(err)
	}

	h := myhttp.NewHandler(ds, l)

	s := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	fmt.Println("Serving...")
	log.Fatal(s.ListenAndServe())
}
