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
	forceTimedMessage = false
)

var (
	lunchTime        = mealTime{hourMinute{12, 15}, hourMinute{13, 45}}
	dinnerTime       = mealTime{hourMinute{17, 45}, hourMinute{19, 15}}
	interval   uint8 = 45
)

type config struct {
	admin      string               // admin user
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
	if value, success := os.LookupEnv(env); success {
		return value
	}

	return def
}

// Initialiser for default port
func getPort(env string, def uint) uint {
	value, success := os.LookupEnv(env)
	if !success {
		return def
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return uint(port)
}

func init() {
	cfg.certPath = getConfigValue("SSL_CERT_PATH", "")
	cfg.keyPath = getConfigValue("SSL_KEY_PATH", "")
	cfg.userBucket = getConfigValue("USER_BUCKET", defaultUserBucket)
	cfg.port = getPort("PORT", 8080)

	// Initialiser variables for other Config members
	accessToken := getConfigValue("FACEBOOK_ACCESS_TOKEN", "")
	dbPath := getConfigValue("CHUMENU_DB_PATH", "test/chumenu.db")

	// Debug Logger
	cfg.debug = log.New(os.Stdout, "", log.Lshortfile)

	// Facebook Send Client
	cfg.sendClient = &facebook.SendClient{AccessToken: accessToken, BaseURL: facebook.APIBase, Metadata: "Churchill Menus"}

	// Facebook Webhook
	cfg.webhook = &facebook.Webhook{AppSecret: getConfigValue("FACEBOOK_APP_SECRET", ""), VerifyToken: getConfigValue("FACEBOOK_VERIFICATION_TOKEN", ""), Handler: eventHandler{commandPrefix: getConfigValue("COMMAND_PREFIX", "/")}, Debug: cfg.debug}

	// Admin User
	cfg.admin = getConfigValue("ADMIN_USER", "")

	// Database Initialisation
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalln(err)
	}
	cfg.db = db

	err = cfg.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(cfg.userBucket)) // nolint: vetshadow
		return err
	})

	if err != nil {
		log.Fatalln(err)
	}

	// Lunch
	gocron.Every(1).Day().At(lunchTime.Start.Before(interval).String()).Do(timedMessage, true, forceTimedMessage)
	// Dinner
	gocron.Every(1).Day().At(dinnerTime.Start.Before(interval).String()).Do(timedMessage, false, forceTimedMessage)

	// api handler
	http.HandleFunc("/webhook", cfg.webhook.ResponseHandler)
	// privacy page
	http.Handle("/privacy", http.FileServer(http.Dir(getConfigValue("PUBLIC_DIR", "public"))))
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
