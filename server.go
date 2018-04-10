package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"strconv"

	"github.com/boltdb/bolt"
	"github.com/jasonlvhit/gocron"
	"github.com/ratorx/chumenu-go/facebook"
)

const (
	defaultUserBucket = "users"
	brunchTime        = "10:30"
	lunchTime         = "11:40"
	dinnerTime        = "17:00"
	forceTimedMessage = false
)

type config struct {
	certPath   string               // path to cert.pem
	sendClient *facebook.SendClient // api client for sending messages
	webhook    *facebook.Webhook    // Facebook Webhook handler
	db         *bolt.DB             // db reference
	keyPath    string               // path to privkey.pem
	port       uint                 // server port
	userBucket string               // bucket for users
	debug      *log.Logger          // Logger for all packages
}

var cfg config

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
	cfg.certPath = getConfigValue("SSL_CERT_PATH", "")
	cfg.keyPath = getConfigValue("SSL_KEY_PATH", "")
	cfg.userBucket = getConfigValue("USER_BUCKET", defaultUserBucket)
	cfg.port = getPort("PORT", 0)

	// Initialiser variables for other Config members
	accessToken := getConfigValue("FACEBOOK_ACCESS_TOKEN", "")
	dbPath := getConfigValue("CHUMENU_DB_PATH", "test/chumenu.db")

	// Debug Logger
	cfg.debug = log.New(os.Stdout, "", log.Lshortfile)

	// Facebook Send Client
	cfg.sendClient = &facebook.SendClient{AccessToken: accessToken, BaseURL: facebook.APIBase, Metadata: "Churchill Menus"}

	// Facebook Webhook
	cfg.webhook = &facebook.Webhook{AppSecret: getConfigValue("FACEBOOK_APP_SECRET", ""), VerifyToken: getConfigValue("FACEBOOK_VERIFICATION_TOKEN", ""), Handler: EventHandler{commandPrefix: getConfigValue("COMMAND_PREFIX", "/")}, Debug: cfg.debug}

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
	// Fire more than necessary because of library bugs
	// Brunch
	gocron.Every(1).Day().At(brunchTime).Do(timedMessage, true, true, forceTimedMessage)
	// Lunch
	gocron.Every(1).Day().At(lunchTime).Do(timedMessage, true, false, forceTimedMessage)
	// Dinner
	gocron.Every(1).Day().At(dinnerTime).Do(timedMessage, false, false, forceTimedMessage)

	// api handler
	http.HandleFunc("/webhook", cfg.webhook.ResponseHandler)
	// privacy page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, getConfigValue("PUBLIC_DIR", "public"))
	})
}

func main() {
	log.SetFlags(0)
	defer cfg.db.Close() // nolint: errcheck
	// start timed messages
	go func() { <-gocron.Start() }()

	// start webserver in default thread
	if cfg.certPath == "" || cfg.keyPath == "" {
		log.Printf("Listening on %v (HTTP)", cfg.port)
		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%v", cfg.port), nil))
	} else {
		log.Printf("Listening on %v (HTTPS)", cfg.port)
		log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%v", cfg.port), cfg.certPath, cfg.keyPath, nil))
	}
}
