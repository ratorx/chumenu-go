package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"io/ioutil"

	"github.com/boltdb/bolt"
	"github.com/jasonlvhit/gocron"
	"github.com/ratorx/chumenu-go/facebook"
)

const (
	defaultPort = 5001
	userBucket  = "users"
)

type config struct {
	appSecret   string
	certPath    string
	keyPath     string
	db          *bolt.DB
	port        uint
	verifyToken string
	client      *facebook.Client
}

var cfg config

func verifyWebhook(response http.ResponseWriter, request *http.Request) {
	urlQueries := request.URL.Query()

	switch {
	case urlQueries.Get("hub.mode") != "subscribe":
		log.Println("verification failed: hub.mode is not subscribe")
		http.Error(response, "hub.mode is not subscribe", http.StatusBadRequest)
	case urlQueries.Get("hub.verify_token") != cfg.verifyToken:
		log.Println("verification failed: invalid hub.verify_token")
		http.Error(response, "invalid hub.verify_token", http.StatusUnauthorized)
	default:
		fmt.Fprintf(response, urlQueries.Get("hub.challenge"))
	}
}

func handler(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		verifyWebhook(response, request)
	case "POST":
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Println(err)
			http.Error(response, "Request Body could not be parsed", http.StatusBadRequest)
			return
		}

		if !checkSHA(request.Header.Get("X-Hub-Signature"), body) {
			log.Println("Invalid SHA1 key in header")
			http.Error(response, "SHA1 validation failed", http.StatusUnauthorized)
			return
		}

		go webhook(body)
		response.WriteHeader(http.StatusOK)
	default:
		log.Printf("%v request", request.Method)
		http.Error(response, "HTTP method not GET or POST", http.StatusMethodNotAllowed)
	}
}

func getConfigValue(env string, def string) string {
	value, success := os.LookupEnv(env)
	if !success {
		return def
	}

	return value
}

func init() {
	log.SetPrefix("chumenu: ")

	cfg.appSecret = getConfigValue("FACEBOOK_APP_SECRET", "")
	cfg.certPath = getConfigValue("SSL_CERT_PATH", "~/.config/chumenu/fullchain.pem")
	cfg.keyPath = getConfigValue("SSL_KEY_PATH", "~/.config/chumenu/privkey.pem")
	cfg.port = defaultPort
	cfg.verifyToken = getConfigValue("FACEBOOK_VERIFICATION_TOKEN", "")

	// Initialiser variables for other Config members
	accessToken := getConfigValue("FACEBOOK_ACCESS_TOKEN", "")
	dbPath := getConfigValue("CHUMENU_DB_PATH", "~/.config/chumenu/chumenu.db")

	// Facebook Client
	cfg.client = &facebook.Client{AccessToken: accessToken}

	// Database Initialisation
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalln(err)
	}
	cfg.db = db

	err = cfg.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(userBucket))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	// Cron Setup
	gocron.Every(1).Day().At("11:40").Do(func() { timedMessage(true, false) })
	gocron.Every(1).Day().At("17:00").Do(func() { timedMessage(false, false) })

	// Web Handler Setup
	http.HandleFunc("/webhook", handler)
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "public/privacy.html") })
}

func main() {
	defer cfg.db.Close() // nolint: errcheck
	// Cron Initialisation
	go func() { <-gocron.Start() }()

	// Webserver Start
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%v", cfg.port), cfg.certPath, cfg.keyPath, nil))
}
