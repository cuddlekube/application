package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var Ver = "1.0.0"
var SHA = "a1b2c3def"

type app struct {
	AppName []appInfo `json:"validate-api"`
}
type appInfo struct {
	Version       string `json:"version"`
	LastCommitSHA string `json:"lastcommitsha"`
}

type helloWorld struct {
	Msg string `json:"message"`
}

func main() {
	log.Print("starting the api")
	r := mux.NewRouter()
	r.HandleFunc("/", root)
	r.HandleFunc("/health", health)

	s := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}

func health(w http.ResponseWriter, r *http.Request) {
	info := appInfo{
		Version:       Ver,
		LastCommitSHA: SHA,
	}

	myApp := app{
		AppName: []appInfo{
			info,
		},
	}

	infoJSON, err := json.Marshal(myApp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(infoJSON)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	hw, err := json.Marshal(helloWorld{"Hello World!"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = w.Write(hw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
