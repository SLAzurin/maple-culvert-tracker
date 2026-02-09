package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	urlFlag := flag.String("url", "", "host name to healthcheck")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatalln("url flag empty")
	}
	retry := true
	for retry {
		retry = false
		r, err := http.Get(*urlFlag)
		if err != nil {
			log.Fatalln(err.Error())
		}
		if r.StatusCode == http.StatusServiceUnavailable {
			retry = true
			dStr := r.Request.Response.Header.Get("Retry-After")
			retryAfterD, err := time.Parse(time.RFC1123, dStr)
			if dStr == "" || err != nil {
				log.Fatal("failed StatusServiceUnavailable invalid date in Retry-After header")
			}
			time.Sleep(time.Until(retryAfterD))
		} else if r.StatusCode != 200 {
			log.Fatalln("failed with " + strconv.Itoa(r.StatusCode))
		}
	}
}
