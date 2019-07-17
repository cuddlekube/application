package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var Ver = "1.0.0"
var SHA = "a1b2c3def"
var TableName = "cuddlykube"
var dynamo *dynamodb.DynamoDB
var local bool
var dynamoURL string

type app struct {
	AppName []appInfo `json:"feed-api"`
}
type appInfo struct {
	Version       string `json:"version"`
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

func init() {
	flag.BoolVar(&local, "local", false, "boolean if set to true will expect dynamo to be available locally")
	flag.StringVar(&dynamoURL, "endpoint-url", "localhost", "default is localhost:8000 override with flag")
	flag.Parse()
}

func main() {
	log.Print("initialising dynamodb")

	config := &aws.Config{
		Region: aws.String("ap-southeast-2"),
	}
	if local {
		log.Print("connecting to local dynamodb")
		config.Endpoint = aws.String(fmt.Sprintf("http://%s:8000", dynamoURL))
		config.Credentials = credentials.NewStaticCredentials("123", "123", "")
	}

	sess := session.Must(session.NewSession(config))
	dynamo = dynamodb.New(sess)

	log.Print("starting the api")

	r := mux.NewRouter()
	r.HandleFunc("/", feed).Methods(http.MethodPost)
	r.HandleFunc("/health", health).Methods(http.MethodGet)

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
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, infoJSON)
}

func feed(w http.ResponseWriter, r *http.Request) {
	var ck cuddlyKube

	// unmarshal the request body into cuddly kube object
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ck); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	defer r.Body.Close()

	// marshal cuddly kube key into the dyanmodb attribute value
	key, err := dynamodbattribute.MarshalMap(ck.CKID)
	if err != nil {
		msg := fmt.Sprintf("error marshaling item into attribute value, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}
	// create putItemInput like every freaking thing in aws go sdk
	input := &dynamodb.UpdateItemInput{
		UpdateExpression: aws.String("SET happiness := 2"),
		Key:              key,
		TableName:        aws.String(TableName),
	}

	// call the put item api
	_, err = dynamo.UpdateItem(input)
	if err != nil {
		msg := fmt.Sprintf("error putting item into cuddlykube table, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": fmt.Sprintf("item %s, is registered", ck.Name)})
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
