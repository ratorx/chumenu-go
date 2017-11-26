package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"io/ioutil"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/jasonlvhit/gocron"
	"github.com/ratorx/chumenu-go/facebook"
)

type config struct {
	appSecret     string           // used for request checksum verification
	certPath      string           // path to cert.pem
	client        *facebook.Client // api client for sending messages
	commandPrefix string           // prefix for testing commands
	db            *bolt.DB         // db reference
	keyPath       string           // path to privkey.pem
	port          uint             // server port
	userBucket    string           // bucket for users
	verifyToken   string           // token used for initial verification
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

		// verify origin of request
		if !checkSHA(request.Header.Get("X-Hub-Signature"), body) {
			log.Println("checksum failed: checksum of body is invalid")
			http.Error(response, "SHA1 validation failed", http.StatusUnauthorized)
			return
		}

		// handle webhook in separate thread to maintain low request time
		go webhook(body)
		response.WriteHeader(http.StatusOK)
	default:
		log.Printf("unknown request type: %v", request.Method)
		http.Error(response, "HTTP method not GET or POST", http.StatusMethodNotAllowed)
	}
}

// Config values with default initialiser
func getConfigValue(env string, def string) string {
	value, success := os.LookupEnv(env)
	if !success {
		return def
	}

	return value
}

// Initialiser for default port
func getPort(env string, def uint) uint {
	value, success := os.LookupEnv(env)
	if !success {
		return def
	}

	port, _ := strconv.Atoi(value)
	return uint(port)
}

func init() {
	cfg.appSecret = getConfigValue("FACEBOOK_APP_SECRET", "")
	cfg.certPath = getConfigValue("SSL_CERT_PATH", "~/.config/chumenu/fullchain.pem")
	cfg.commandPrefix = getConfigValue("COMMAND_PREFIX", "/")
	cfg.keyPath = getConfigValue("SSL_KEY_PATH", "~/.config/chumenu/privkey.pem")
	cfg.userBucket = getConfigValue("USER_BUCKET", "")
	cfg.verifyToken = getConfigValue("FACEBOOK_VERIFICATION_TOKEN", "")
	cfg.port = getPort("PORT", 5001)

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
		_, err := tx.CreateBucketIfNotExists([]byte(cfg.userBucket))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	// timed message for subscribers
	gocron.Every(1).Day().At("11:40").Do(func() { timedMessage(true, false) })
	gocron.Every(1).Day().At("17:00").Do(func() { timedMessage(false, false) })

	// api handler
	http.HandleFunc("/webhook", handler)
	// privacy page
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "public/privacy.html") })
}

func main() {
	log.SetFlags(0)
	defer cfg.db.Close() // nolint: errcheck
	// start timed messages
	go func() { <-gocron.Start() }()

	// start webserver in default thread
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%v", cfg.port), cfg.certPath, cfg.keyPath, nil))
}
