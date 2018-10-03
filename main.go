package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func serveStatic(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := ioutil.ReadFile("public/" + path)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		http.ServeContent(w, r, path, time.Now(), bytes.NewReader(f))
	}
}

//ResponseURL response
type ResponseURL struct {
	Hash  string
	Value string
	Url   string
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.Itoa(int(h.Sum32()))
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["url"]

	if !ok || len(keys) <= 0 {
		http.Error(w, "Could not find url.", http.StatusBadRequest)
		return
	}

	url := keys[0]
	hash := hash(url)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if HashMap[hash] == "" {
		HashMap[hash] = url
	}

	json.NewEncoder(w).Encode(&ResponseURL{url, hash, "/s/" + hash})
}

func redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	if len(HashMap[hash]) > 0 {
		w.Header().Set("Location", HashMap[hash])
		w.WriteHeader(http.StatusPermanentRedirect)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Could not find the URL found.")
}

//HashMap the conversion between hash and value
var HashMap = make(map[string]string)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/shorten", shortenURL)
	router.HandleFunc("/s/{hash:.*}", redirect)
	router.HandleFunc("/", serveStatic("index.html"))
	http.Handle("/", router)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
