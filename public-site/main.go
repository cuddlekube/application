package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const STATIC_DIR = "/img/"

var listAPIURL = "http://list-api%s:8080"
var feedAPIURL = "http://feed-api%s:8080"
var internalDomain = ""

var Ver = "1.0.0"
var SHA = "a1b2c3def"

// cuddly kube that matches the cuddly kube table
//- ckid -- HASH
//- name -- String
//- type -- String (aws server classes?)
//- service -- int (e.g 20 years in service)
//- happiness -- int (1 being shit 10 being super happy)
//- petname -- String
//- os -- String (linux, windows)
//- image -- String
type cuddlyKube struct {
	CKID      string `json:"ckid"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Service   int    `json:"service"`
	Happiness int    `json:"happiness"`
	Petname   string `json:"petname"`
	OS        string `json:"os"`
	Image     string `json:"image"`
}

type app struct {
	AppName []appInfo `json:"public-site"`
}
type appInfo struct {
	Version       string `json:"version"`
	LastCommitSHA string `json:"lastcommitsha"`
}

type helloWorld struct {
	Msg string `json:"message"`
}

func main() {
	if os.Getenv("INTERNAL_DOMAIN") != "" {
		internalDomain = "." + os.Getenv("INTERNAL_DOMAIN")
	}
	feedAPIURL = fmt.Sprintf(feedAPIURL, internalDomain)
	listAPIURL = fmt.Sprintf(listAPIURL, internalDomain)
	log.Print("starting the api")
	r := mux.NewRouter().StrictSlash(true)
	r.PathPrefix(STATIC_DIR).
		Handler(http.StripPrefix(STATIC_DIR, http.FileServer(http.Dir("."+STATIC_DIR))))
	r.HandleFunc("/", root)
	r.HandleFunc("/list", list)
	r.HandleFunc("/register", register)
	r.HandleFunc("/health", version)

	s := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}

func version(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	page := template.Must(template.ParseFiles("tmpl/homepage.html"))

	data := "nothing"
	err := page.Execute(w, data)
	fmt.Println(err)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	page, err := template.New("list.html").Funcs(template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }}).ParseFiles("tmpl/list.html")
	if err != nil {
		log.Fatal(err)
	}

	// res, err := http.Get(listAPIURL)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer res.Body.Close()
	// if res.StatusCode != 200 {
	// 	log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	// }

	mock := `[
		{
			"ckid": "1234",
			"name": "My first server",
			"type": "x1.32xlarge",
			"service": 2,
			"happiness": 4,
			"petname": "Titan",
			"os": "Linux",
			"image": "/img/x1.png"
		},
		{
			"ckid": "2345",
			"name": "Real ARM server",
			"type": "a1.large",
			"service": 2,
			"happiness": 4,
			"petname": "Tiny",
			"os": "Linux",
			"image": "/img/a1.png"
		}
	]`
	parsedServers := []cuddlyKube{}
	if err := json.Unmarshal([]byte(mock), &parsedServers); err != nil {
		panic(err)
	}

	data := struct {
		Servers []cuddlyKube
	}{}
	data.Servers = parsedServers

	err = page.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	page := template.Must(template.ParseFiles("tmpl/register.html"))

	data := "nothing"
	err := page.Execute(w, data)
	fmt.Println(err)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
