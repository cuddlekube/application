package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

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

type app struct {
	AppName []appInfo `json:"list-api"`
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
	flag.StringVar(&dynamoURL, "url", "http://localhost:8000", "default is localhost:8000 override with flag")
	flag.BoolVar(&local, "local", false, "boolean if set to true will expect dynamo to be available locally")
	flag.Parse()

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
	xray.AWS(dynamo.Client)

	log.Print("starting the api")

	r := mux.NewRouter()
	r.HandleFunc("/", list).Methods(http.MethodGet)
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

// list scans the dynamo db table and returns all result
// this can be potentially spammed
// look into query with paging etc
func list(w http.ResponseWriter, r *http.Request) {

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("REQUEST_DUMP: %s\n", string(requestDump))

	attributes := make(map[string]*string)
	attributes["#v"] = aws.String("image")
	i := &dynamodb.ScanInput{
		TableName:                aws.String(TableName),
		FilterExpression:         aws.String("size(#v) > :num"),
		ExpressionAttributeNames: attributes,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":num": {
				N: aws.String("0"),
			}},
	}
	o, err := dynamo.ScanWithContext(r.Context(), i)
	if err != nil {
		msg := fmt.Sprintf("error getting items from cuddlykube table, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	cks := make(map[int]cuddlyKube)
	for k, v := range o.Items {
		var ck cuddlyKube
		err = dynamodbattribute.UnmarshalMap(v, &ck)
		if err != nil {
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
