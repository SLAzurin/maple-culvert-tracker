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

	r, err := http.NewRequest("GET", *urlFlag, nil)
	if r.Response.StatusCode != 200 || err != nil {
		log.Fatalln("failed with " + strconv.Itoa(r.Response.StatusCode) + err.Error())
	}
}
