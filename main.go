package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

var (
	serverlog = flag.String(
		"log",
		"",
		"specify a log file to write server logs to",
	)
)

func main() {
	flag.Parse()

	if *serverlog != "" {
		f, err := os.OpenFile(*serverlog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		log.SetOutput(f)
	}

	var routes = Routes{
		Route{
			Name:        "Index",
			Method:      "GET",
			Pattern:     "/",
			HandlerFunc: Index,
		},
		Route{
			Name:        "NewVolume",
			Method:      "POST",
			Pattern:     "/catalog",
			HandlerFunc: Catalog,
		},
		Route{
			Name:        "GetCatalog",
			Method:      "GET",
			Pattern:     "/catalog",
			HandlerFunc: Catalog,
		},
		Route{
			Name:        "GetVolume",
			Method:      "GET",
			Pattern:     "/catalog/{volID}",
			HandlerFunc: CatalogSingle,
		},
		Route{
			Name:        "UpdateVolume",
			Method:      "POST",
			Pattern:     "/catalog/{volID}",
			HandlerFunc: CatalogSingle,
		},
	}

	http.ListenAndServe(":8080", NewRouter(routes))
}
