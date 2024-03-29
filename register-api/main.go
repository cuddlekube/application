package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

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
	AppName []appInfo `json:"register-api"`
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
	xray.AWS(dynamo.Client)

	log.Print("starting the api")

	r := mux.NewRouter()
	r.HandleFunc("/", register).Methods(http.MethodPost)
	r.HandleFunc("/health", health).Methods(http.MethodGet)

	s := &http.Server{
		Handler:      xray.Handler(xray.NewFixedSegmentNamer("register-api"), r),
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
	ck.CKID = strconv.Itoa(rand.Intn(10000000))

	valid, err := validate(ck, r)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	if !valid {
		msg := fmt.Sprintf("please ensure that the cuddly kube object has Name, OS and Type values set")
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}

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
	_, err = dynamo.PutItemWithContext(r.Context(), input)
	if err != nil {
		msg := fmt.Sprintf("error putting item into cuddlykube table, %s ", err.Error())
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": fmt.Sprintf("item %s, is registered", ck.Name)})
}

func validate(ck cuddlyKube, r *http.Request) (bool, error) {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	client := xray.Client(&http.Client{Transport: tr})
	buf, err := json.Marshal(ck)
	if err != nil {
		return false, fmt.Errorf("unable to marshal ck into byte slice: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, validateURL, bytes.NewBuffer(buf))
	if err != nil {
		return false, fmt.Errorf("error creating new http request, %s ", err.Error())
	}

	log.Printf("new http request created")

	resp, err := client.Do(req.WithContext(r.Context()))
	if err != nil {
		return false, fmt.Errorf("error calling the validate api, %s ", err.Error())
	}

	log.Printf("called the validate-api to validate ck: %s", ck.CKID)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var v result
	err = json.Unmarshal(body, &v)
	if err != nil {
		return false, fmt.Errorf("error unmarshaling validate-api response %s ", err.Error())
	}

	log.Printf("resp body %v", v)

	return v.Valid, nil
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
