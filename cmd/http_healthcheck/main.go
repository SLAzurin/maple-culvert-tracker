package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

func main() {
	urlFlag := flag.String("url", "", "host name to healthcheck")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatalln("url flag empty")
	}

	r, err := http.Get(*urlFlag)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if r.StatusCode != 200 {
		log.Fatalln("failed with " + strconv.Itoa(r.StatusCode))
	}
}
