package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)


var Ver = "1.0.0"
var SHA = "a1b2c3def"
var TableName = "cuddlykube"
var dynamo *dynamodb.DynamoDB
var local bool

type app struct {
	AppName []appInfo `json:"list-api"`
}
type appInfo struct {
	Version       string `json:"version"`
	LastCommitSHA string `json:"lastcommitsha"`
}


// Struct to return data
// TODO: add fields that are relevant for the example
type cuddlyKube struct {
	CKID string `json:"ckid"`
	Name string `json:"name"`
}

func init() {
	flag.BoolVar(&local, "local", false, "boolean if set to true will expect dynamo to be available at localhost:8000 ")
	flag.Parse()
}


func main() {

	log.Print("initialising dynamodb")

	config := &aws.Config{
		Region: aws.String("ap-southeast-2"),
	}
	if local {
		log.Print("connecting to local dynamodb")
		config.Endpoint = aws.String("http://localhost:8000")
		config.Credentials = credentials.NewStaticCredentials("123", "123", "")
	}

	sess := session.Must(session.NewSession(config))
	dynamo = dynamodb.New(sess)

	log.Print("starting the api")

	r := mux.NewRouter()
	r.HandleFunc("/", list).Methods(http.MethodGet)
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

// list scans the dynamo db table and returns all result
// this can be potentially spammed
// look into query with paging etc
func list(w http.ResponseWriter, r *http.Request) {

	i := &dynamodb.ScanInput{
		TableName: aws.String(TableName),
	}
	o, err := dynamo.Scan(i)
	if err != nil {
		msg := fmt.Sprintf("error putting item into cuddlykube table, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	cks := make(map[int]cuddlyKube)
	for k, v := range o.Items {
		var ck cuddlyKube
		err = dynamodbattribute.UnmarshalMap(v, &ck)
		if err  != nil {
			msg := fmt.Sprintf("error unmarshaling item into cuddlykube, %s ", err.Error())
			log.Println(msg)
			continue
		}
		cks[k] = ck
	}
	respondWithJSON(w, http.StatusOK, cks)
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
