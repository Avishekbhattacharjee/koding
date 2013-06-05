package main

import (
	"github.com/gorilla/mux"
	"io"
	"koding/kontrol/kontroldaemon/clientconfig"
	"koding/kontrol/kontroldaemon/workerconfig"
	"koding/kontrol/kontrolproxy/proxyconfig"
	"koding/tools/config"
	"log"
	"net/http"
	"strconv"
)

type ProxyPostMessage struct {
	Name      string
	Username  string
	Domain    string
	Mode      string
	Key       string
	RabbitKey string
	Host      string
	Hostdata  string
}

var clientDB *clientconfig.ClientConfig
var kontrolConfig *workerconfig.WorkerConfig
var proxyDB *proxyconfig.ProxyConfiguration
var amqpWrapper *AmqpWrapper

func init() {
	log.SetPrefix("kontrol-api ")
}

func main() {
	amqpWrapper = setupAmqp()

	var err error
	kontrolConfig, err = workerconfig.Connect()
	if err != nil {
		log.Fatalf("wokerconfig mongodb connect: %s", err)
	}

	proxyDB, err = proxyconfig.Connect()
	if err != nil {
		log.Fatalf("proxyconfig mongodb connect: %s", err)
	}

	clientDB, err = clientconfig.Connect()
	if err != nil {
		log.Fatalf("proxyconfig mongodb connect: %s", err)
	}

	port := strconv.Itoa(config.Current.Kontrold.Api.Port)

	rout := mux.NewRouter()
	rout.HandleFunc("/", home).Methods("GET")

	// Deployment handlers
	rout.HandleFunc("/deployments", GetClients).Methods("GET")
	rout.HandleFunc("/deployments", CreateClient).Methods("POST")
	rout.HandleFunc("/deployments/{build}", GetClient).Methods("GET")

	// Worker handlers
	rout.HandleFunc("/workers", GetWorkers).Methods("GET")
	rout.HandleFunc("/workers/{uuid}", GetWorker).Methods("GET")
	rout.HandleFunc("/workers/{uuid}/{action}", UpdateWorker).Methods("PUT")
	rout.HandleFunc("/workers/{uuid}", DeleteWorker).Methods("DELETE")

	// Proxy handlers
	rout.HandleFunc("/proxies", GetProxies).Methods("GET")
	rout.HandleFunc("/proxies/{proxyname}", GetProxy).Methods("GET")
	rout.HandleFunc("/proxies/{proxyname}", CreateProxy).Methods("POST")
	rout.HandleFunc("/proxies/{proxyname}", DeleteProxy).Methods("DELETE")

	// Service handlers
	rout.HandleFunc("/services", GetProxyUsers).Methods("GET")
	rout.HandleFunc("/services/{username}", GetProxyServices).Methods("GET")
	rout.HandleFunc("/services/{username}", CreateProxyUser).Methods("POST")
	rout.HandleFunc("/services/{username}/{servicename}", GetKeyList).Methods("GET")
	rout.HandleFunc("/services/{username}/{servicename}", CreateProxyService).Methods("POST")
	rout.HandleFunc("/services/{username}/{servicename}", DeleteProxyService).Methods("DELETE")
	rout.HandleFunc("/services/{username}/{servicename}/{key}", GetKey).Methods("GET")
	rout.HandleFunc("/services/{username}/{servicename}/{key}", DeleteKey).Methods("DELETE")

	// Domain handlers
	rout.HandleFunc("/domains", GetDomains).Methods("GET")
	rout.HandleFunc("/domains/{domain}", GetDomain).Methods("GET")
	rout.HandleFunc("/domains/{domain}", CreateDomain).Methods("POST")
	rout.HandleFunc("/domains/{domain}", DeleteDomain).Methods("DELETE")

	// Rule handlers
	rout.HandleFunc("/rules", GetRules).Methods("GET")
	rout.HandleFunc("/rules/{username}", GetRulesServices).Methods("GET")
	rout.HandleFunc("/rules/{username}/{servicename}", GetRule).Methods("GET")
	rout.HandleFunc("/rules/{username}/{servicename}", CreateRule).Methods("POST")

	// Statistics handlers
	rout.HandleFunc("/stats", GetStats).Methods("GET")
	rout.HandleFunc("/stats", DeleteStats).Methods("DELETE")
	rout.HandleFunc("/stats/domains", GetDomainStats).Methods("GET")
	rout.HandleFunc("/stats/domains/{domain}", GetSingleDomainStats).Methods("GET")
	rout.HandleFunc("/stats/proxies", GetProxyStats).Methods("GET")
	rout.HandleFunc("/stats/proxies/{proxy}", GetSingleProxyStats).Methods("GET")

	// Rollbar api
	rout.HandleFunc("/rollbar", rollbar).Methods("POST")

	log.Printf("kontrol api is started. serving at :%s ...", port)

	http.Handle("/", rout)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Println(err)
	}
}

func home(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "Hello world - kontrol api!\n")
}
