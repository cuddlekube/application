package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const imageDir = "/img/"

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
	r.PathPrefix(imageDir).
		Handler(http.StripPrefix(imageDir, http.FileServer(http.Dir("."+imageDir))))
	r.HandleFunc("/", root)
	r.HandleFunc("/list", list)
	r.HandleFunc("/register", register).Methods(http.MethodGet)
	r.HandleFunc("/register", sendregistration).Methods(http.MethodPost)
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
		log.Print(err)
	}

	serviceclient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, listAPIURL, nil)
	if err != nil {
		log.Print("Request error")
		log.Print(err)
	}
	res, getErr := serviceclient.Do(req)
	if getErr != nil {
		log.Print("Do error")
		log.Print(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Print("Read error")
		log.Print(res.Body)
		log.Print(string(body))
		log.Print(readErr)
	}

	parsedServers := make(map[int]cuddlyKube)
	if err := json.Unmarshal([]byte(body), &parsedServers); err != nil {
		log.Print(err)
	}

	data := struct {
		Servers map[int]cuddlyKube
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

func sendregistration(w http.ResponseWriter, r *http.Request) {
	// nothing ye
}
