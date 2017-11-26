package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/boltdb/bolt"
	"github.com/ratorx/chumenu-go/menus"
)

const (
	// Subscribe Messages
	subscribeSuccess = "Subscribed to receive messages."
	subscribeFail    = "Currently subscribed."

	// Unsubscribe Messages
	unsubscribeSuccess = "Unsubscribed from receiving messages."
	unsubscribeFail    = "Not currently subscribed."

	// Other defaults
	help         = "Available commands:\n*subscribe* - Receive regular menu updates\n*unsubscribe* - Unsubscribe from menu updates\n*lunch* - Get the next lunch menu\n*dinner* - Get the next the dinner menu"
	unrecognised = "Command not recognised. Type help for a list of available commands."
	unexpected   = "Unexpected Error. Will fix ASAP."
)

func textMessageWithLog(sender string, message string) {
	err := cfg.client.TextMessage(sender, message)
	if err != nil {
		log.Println(err)
	}
}

func subscribeHandler(sender string) {
	s := []byte(sender)

	err := cfg.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cfg.userBucket))
		if b == nil {
			return fmt.Errorf("database corrupted: bucket %v not found", cfg.userBucket)
		}

		v := b.Get(s)
		if v != nil {
			go textMessageWithLog(sender, subscribeFail)
			return nil
		}

		err := b.Put(s, []byte{})
		if err != nil {
			go textMessageWithLog(sender, unexpected)
			return err
		}
		go textMessageWithLog(sender, subscribeSuccess)

		return nil
	})

	if err != nil {
		log.Println(err)
	}

}

func unsubscribeHandler(sender string) {
	s := []byte(sender)

	err := cfg.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cfg.userBucket))
		if b == nil {
			return fmt.Errorf("database corrupted: bucket %v not found", cfg.userBucket)
		}

		v := b.Get(s)
		if v == nil {
			go textMessageWithLog(sender, unsubscribeFail)
			return nil
		}

		err := b.Delete(s)
		if err != nil {
			go textMessageWithLog(sender, unexpected)
			return err
		}
		go textMessageWithLog(sender, unsubscribeSuccess)

		return nil
	})

	if err != nil {
		log.Println(err)
	}
}
func defaultHandler(sender, text string) {
	log.Printf("Unrecognised command \"%v\"", text)
	textMessageWithLog(sender, unrecognised)
}

func parseMessage(sender string, text string) {

	text = strings.TrimSpace(text)
	text = strings.Trim(text, "*_`")
	text = strings.ToLower(text)

	if !strings.HasPrefix(text, cfg.commandPrefix) {
		defaultHandler(sender, text)
		return
	}

	text = strings.TrimPrefix(text, cfg.commandPrefix)

	switch text {
	case "subscribe", "s":
		subscribeHandler(sender)
	case "unsubscribe", "u":
		unsubscribeHandler(sender)
	case "help", "h":
		textMessageWithLog(sender, help)
	case "lunch", "l":
		menuMessage(sender, true, true)
	case "dinner", "d":
		menuMessage(sender, false, true)
	default:
		defaultHandler(sender, text)
	}
}

func webhook(body []byte) {
	type Message struct {
		Text string `json:"text"`
	}

	type Person struct {
		ID string `json:"id"`
	}

	type Messaging struct {
		Sender Person  `json:"sender"`
		M      Message `json:"message"`
	}

	type Entry struct {
		Messages []Messaging `json:"messaging"`
	}

	type Object struct {
		Entries []Entry `json:"entry"`
	}

	o := Object{}
	err := json.Unmarshal(body, &o)
	if err != nil {
		log.Println(err)
	}

	for i := range o.Entries {
		for j := range o.Entries[i].Messages {
			m := &o.Entries[i].Messages[j]
			sender := m.Sender.ID
			text := m.M.Text
			parseMessage(sender, text)
		}
	}
}

func mealMessage(sender string, prefix string, meal menus.Meal) {
	textMessageWithLog(sender, prefix+"\n"+meal.String())
}

// Replace block with cache
func menuMessage(sender string, isLunch bool, forceSend bool) {
	currentTime := time.Now()
	block, _ := menus.GetData(uint8(currentTime.Weekday())) // Normalise to UNIX days
	hour := currentTime.Hour()
	minute := currentTime.Minute()

	var prefix string
	var meal menus.Meal

	if isLunch {
		if hour > 13 || (hour == 13 && minute > 45) {
			prefix = "Tomorrow's Lunch:"
			meal = block.Next.Lunch
		} else {
			prefix = "Today's Lunch:"
			meal = block.Current.Lunch
		}
	} else {
		if hour > 19 || (hour == 19 && minute > 15) {
			prefix = "Tomorrow's Dinner:"
			meal = block.Next.Dinner
		} else {
			prefix = "Today's Dinner:"
			meal = block.Current.Dinner
		}
	}

	if forceSend || meal.String() != " - TBC" {
		mealMessage(sender, prefix, meal)
	}
}

func timedMessage(isLunch, forceSend bool) {
	cfg.db.View(func(tx *bolt.Tx) error { // nolint: errcheck
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(cfg.userBucket))
		log.Println("Entered User Bucket")
		b.ForEach(func(k, v []byte) error { // nolint: errcheck
			go menuMessage(string(k), isLunch, forceSend)
			return nil
		})
		return nil
	})
}
