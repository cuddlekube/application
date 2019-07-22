package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var Ver = "1.0.0"
var SHA = "a1b2c3def"
var TableName = "cuddlykube"
var dynamo *dynamodb.DynamoDB
var local bool
var dynamoURL string
var validateURL = "http://validate-api%s:8080"
var internalDomain string

type app struct {
	AppName []appInfo `json:"validate-api"`
}

type appInfo struct {
	Version       string `json:"health"`
	LastCommitSHA string `json:"lastcommitsha"`
}

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

type result struct {
	Valid   bool     `json:"valid"`
	Message []string `json:"message"`
}

func init() {
	flag.BoolVar(&local, "local", false, "boolean if set to true will expect dynamo to be available at localhost:8000 ")
	flag.StringVar(&dynamoURL, "endpoint-url", "http://localhost:8000", "default is localhost:8000 override with flag")
	flag.Parse()

	if os.Getenv("INTERNAL_DOMAIN") != "" {
		internalDomain = "." + os.Getenv("INTERNAL_DOMAIN")
	}
	validateURL = fmt.Sprintf(validateURL, internalDomain)

}

func main() {
	log.Print("initialising dynamodb")

	config := &aws.Config{
		Region: aws.String("ap-southeast-2"),
	}
	if local {
		log.Print("connecting to local dynamodb")
		config.Endpoint = aws.String(dynamoURL)
		config.Credentials = credentials.NewStaticCredentials("123", "123", "")
	}

	sess := session.Must(session.NewSession(config))
	dynamo = dynamodb.New(sess)

	log.Print("starting the api")

	r := mux.NewRouter()
	r.HandleFunc("/", validate).Methods(http.MethodPost)
	r.HandleFunc("/health", health).Methods(http.MethodGet)

	s := &http.Server{
		Handler:      xray.Handler(xray.NewFixedSegmentNamer("list-api"), r),
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
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, infoJSON)
}

func validate(w http.ResponseWriter, r *http.Request) {
	log.Println("received request for validation")
	var ck cuddlyKube

	// unmarshal the request body into cuddly kube object
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ck); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	defer r.Body.Close()

	var res result
	res.Valid = true

	// validation chain
	if ck.Name == "" {
		res.Valid = false
		res.Message = append(res.Message, "cuddly kube name not set")
	}

	if ck.OS == "" {
		res.Valid = false
		res.Message = append(res.Message, "cuddly kube os not set")
	}

	if ck.Type == "" {
		res.Valid = false
		res.Message = append(res.Message, "cuddly kube type not set")
	}

	log.Printf("cuddly kube validated with result: %s", res.Valid)
	respondWithJSON(w, http.StatusOK, res)
}

// helper for responding with error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// helper for responding with json
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
