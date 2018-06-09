package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	addr     = flag.String("addr", ":8080", "serve http on `address`")
	dataFile = flag.String("f", "redirs.yaml", "the YAML configuration file")
)

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		panic(err)
	}
	h, err := newHandler(data)
	if err != nil {
		panic(err)
	}

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
