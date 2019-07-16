package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

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
	AppName []appInfo `json:"register-api"`
}

type appInfo struct {
	Version       string `json:"health"`
	LastCommitSHA string `json:"lastcommitsha"`
}

// Struct to return data
// TODO: add fields that are relevant for the example

type cuddlyKube struct {
	CKID string `json:"ckid"`
	Name string `json:"name"`
}

func init() {
	//flag.StringVar(&nsSuffixIgnore, "ns-suffix-ignore", "", "Comma separated list of namespaces that should not have the suffix applied (or env NS_SUFFIX_IGNORE)")

	flag.BoolVar(&local, "local", false, "-l ")
	flag.Parse()
}

// main function which is initialise dynamo connection and also the http server
func main() {

	log.Print("initialising dynamo")

	config := &aws.Config{
		Region:      aws.String("ap-southeast-2"),
		Endpoint:    aws.String("http://localhost:8000"),
		Credentials: credentials.NewStaticCredentials("123", "123", ""),
	}
	sess := session.Must(session.NewSession(config))

	dynamo = dynamodb.New(sess)

	log.Print("starting the api")

	r := mux.NewRouter()
	r.HandleFunc("/health", health).Methods(http.MethodGet)
	r.HandleFunc("/", register).Methods(http.MethodPost)

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

func register(w http.ResponseWriter, r *http.Request) {
	var ck cuddlyKube

	// unmarshal the request body into cuddly kube object
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ck); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	defer r.Body.Close()

	// marshal cuddly kube into the dyanmodb attribute value map
	av, err := dynamodbattribute.MarshalMap(ck)
	if err != nil {
		msg := fmt.Sprintf("error marshaling item into attribute value, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}

	// create putItemInput like every freaking thing in aws go sdk
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(TableName),
	}

	// call the put item api
	_, err = dynamo.PutItem(input)
	if err != nil {
		msg := fmt.Sprintf("error putting item into cuddlykube table, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": fmt.Sprintf("item %s, is registered", item.Name)})
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
