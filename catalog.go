package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
)

// Volume is an entry in the catalog
type Volume struct {
	BarCode        string    `json:"barcode" db:"barcode"`
	MediaType      string    `json:"media_type" db:"media_type"`
	Capacity       int64     `json:"capacity" db:"capacity"`
	Used           int64     `json:"used" db:"used"`
	BlockSize      int       `json:"blocksize" db:"blocksize"`
	Priority       int       `json:"priority" db:"priority"`
	LabelTime      time.Time `json:"label_time" db:"label_time"`
	ModifyTime     time.Time `json:"modify_time" db:"modify_time"`
	MountTime      time.Time `json:"mount_time" db:"mount_time"`
	NeedsAttention bool      `json:"needs_attention" db:"needs_attention"`
	InUse          bool      `json:"in_use" db:"in_use"`
	Labeled        bool      `json:"labeled" db:"labeled"`
	MediaBad       bool      `json:"media_bad" db:"bad_media"`
	CleaningMedia  bool      `json:"cleaning_media" db:"cleaning_media"`
	WriteProtect   bool      `json:"write_protect" db:"write_protect"`
	ReadOnly       bool      `json:"readonly" db:"readonly"`
	Draining       bool      `json:"draining" db:"draining"`
	Unavailable    bool      `json:"unavailable" db:"unavailable"`
	Full           bool      `json:"full" db:"full"`
}

var dbmap = initDb()

func initDb() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "catalog_db.bin")
	if err != nil {
		log.Fatalln("sql.Open failed:", err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbmap.AddTableWithName(Volume{}, "catalog").SetKeys(false, "BarCode")

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		log.Fatalln("Create tables failed failed:", err)
	}

	return dbmap
}

// Index handles the top level directory request
// and will be used fro API documentation
func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintln(w, "Document API Here")
}

// Catalog is for handling requests dealing with the entire Catalog
// like listing all volumes
func Catalog(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var volumes []Volume
		_, err := dbmap.Select(&volumes, "SELECT * FROM catalog")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(volumes); err != nil {
			panic(err)
		}
	case "POST":
		var volume Volume
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err = r.Body.Close(); err != nil {
			panic(err)
		}
		if err = json.Unmarshal(body, &volume); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		err = dbmap.Insert(&volume)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(volume); err != nil {
			panic(err)
		}
	case "PUT":
		http.Error(w, "Invalid request method.", 405)
	case "DELETE":
		http.Error(w, "Invalid request method.", 405)
	default:
		http.Error(w, "Invalid request method.", 405)
	}
}

// CatalogSingle will handling create, read, update, delete individual volumes
func CatalogSingle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		vars := mux.Vars(r)
		barcode := vars["barcode"]
		var volume Volume
		err := dbmap.SelectOne(&volume, "SELECT * FROM catalog WHERE barcode=?", barcode)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(volume); err != nil {
			panic(err)
		}
	case "POST":
		var volume Volume
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err = r.Body.Close(); err != nil {
			panic(err)
		}
		if err = json.Unmarshal(body, &volume); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		_, err = dbmap.Update(&volume)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusAccepted)
		if err = json.NewEncoder(w).Encode(volume); err != nil {
			panic(err)
		}
	case "PUT":
		http.Error(w, "Invalid request method.", 405)
	case "DELETE":
		var volume Volume
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err = r.Body.Close(); err != nil {
			panic(err)
		}
		if err = json.Unmarshal(body, &volume); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		_, err = dbmap.Delete(&volume)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusAccepted)
		if err = json.NewEncoder(w).Encode(volume); err != nil {
			panic(err)
		}
	default:
		http.Error(w, "Invalid request method.", 405)
	}
}
