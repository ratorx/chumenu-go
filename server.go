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

const (
	userBucket = "users"
)

type config struct {
	certPath       string               // path to cert.pem
	sendClient     *facebook.SendClient // api client for sending messages
	webhook        *facebook.Webhook    // Facebook Webhook handler
	messageHandler string               // Write a Message Handler
	commandPrefix  string               // prefix for testing commands
	db             *bolt.DB             // db reference
	keyPath        string               // path to privkey.pem
	port           uint                 // server port
	userBucket     string               // bucket for users
	debug          *Log.Logger          // Logger for all packages
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
	cfg.certPath = getConfigValue("SSL_CERT_PATH", "~/.config/chumenu/fullchain.pem")
	cfg.commandPrefix = getConfigValue("COMMAND_PREFIX", "/")
	cfg.keyPath = getConfigValue("SSL_KEY_PATH", "~/.config/chumenu/privkey.pem")
	cfg.userBucket = userBucket
	cfg.port = getPort("PORT", 5001)

	// Initialiser variables for other Config members
	accessToken := getConfigValue("FACEBOOK_ACCESS_TOKEN", "")
	dbPath := getConfigValue("CHUMENU_DB_PATH", "~/.config/chumenu/chumenu.db")

	// Facebook Send Client
	cfg.client = &facebook.SendClient{AccessToken: accessToken, BaseURL: facebook.APIBase, Metadata: "Churchill Menus"}

	// Facebook Webhook
	cfg.webhook = &facebook.Webhook{AppSecret: getConfigValue("FACEBOOK_APP_SECRET", ""), VerifyToken: getConfigValue("FACEBOOK_VERIFICATION_TOKEN", "")}

	// Facebook Event Handler

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
	http.HandleFunc("/webhook", webhook.ResponseHandler)
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
