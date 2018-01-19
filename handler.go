package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ratorx/chumenu-go/facebook"
	"github.com/ratorx/chumenu-go/menus"
	"strings"
	"time"
)

type EventHandler struct {
	commandPrefix string
}

// Defined keywords
const (
	lunch       = "lunch"
	dinner      = "dinner"
	help        = "help"
	subscribe   = "subscribe"
	unsubscribe = "unsubscribe"
)

// Common quick replies
var (
	standardQR       = []facebook.QuickReply{facebook.QuickReply{lunch}, facebook.QuickReply{dinner}, facebook.QuickReply{help}}
	subscriptionQR   = []facebook.QuickReply{facebook.QuickReply{unsubscribe}, facebook.QuickReply{lunch}, facebook.QuickReply{dinner}, facebook.QuickReply{help}}
	unsubscriptionQR = []facebook.QuickReply{facebook.QuickReply{subscribe}, facebook.QuickReply{help}}
	helpQR           = []facebook.QuickReply{facebook.QuickReply{lunch}, facebook.QuickReply{dinner}, facebook.QuickReply{subscribe}, facebook.QuickReply{unsubscribe}}
	defQR            = []facebook.QuickReply{facebook.QuickReply{help}}
)

// standard Messages
const (
	// Subscribe Messages
	subscribeSuccess = "Subscribed to receive messages."
	subscribeFail    = "Currently subscribed."

	// Unsubscribe Messages
	unsubscribeSuccess = "Unsubscribed from receiving messages."
	unsubscribeFail    = "Not currently subscribed."

	// Other defaults
	helpMessage  = "Available commands:\n*subscribe* - Receive regular menu updates\n*unsubscribe* - Unsubscribe from menu updates\n*lunch* - Get the next lunch menu\n*dinner* - Get the next the dinner menu"
	unrecognised = "Command not recognised. Type help for a list of available commands."
	unexpected   = "Unexpected Error. Will fix ASAP."
)

func responseMessage(r string, text string, qr []facebook.QuickReply) {
	err := cfg.sendClient.SendMessage(r, text, facebook.Response, qr)
	if err != nil {
		cfg.debug.Print(err)
	}
}

func subscriptionMessage(r string, text string, qr []facebook.QuickReply) {
	err := cfg.sendClient.SendMessage(r, text, facebook.Subscription, qr)
	if err != nil {
		cfg.debug.Print(err)
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
			go responseMessage(sender, subscribeFail, subscriptionQR)
			return nil
		}

		err := b.Put(s, []byte{})
		if err != nil {
			go responseMessage(sender, unexpected, standardQR)
			return err
		}
		go responseMessage(sender, subscribeSuccess, subscriptionQR)

		return nil
	})

	if err != nil {
		cfg.debug.Print(err)
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
			go responseMessage(sender, unsubscribeFail, unsubscriptionQR)
			return nil
		}

		err := b.Delete(s)
		if err != nil {
			go responseMessage(sender, unexpected, standardQR)
			return err
		}
		go responseMessage(sender, unsubscribeSuccess, unsubscriptionQR)

		return nil
	})

	if err != nil {
		cfg.debug.Println(err)
	}
}

func getMenu(isLunch bool) (string, menus.Meal) {
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

	return prefix, meal
}

func menuMessage(r string, isLunch bool) {
	prefix, meal := getMenu(isLunch)
	responseMessage(r, prefix+"\n"+meal.String(), standardQR)
}

func timedMessage(isLunch, forceSend bool) {
	prefix, meal := getMenu(isLunch)

	if !forceSend && meal.String() == " - TBC" {
		cfg.debug.Print("data unavailable for timed message")
		return
	}

	menu := prefix + "\n" + meal.String()

	var num uint
	var mealName string

	if isLunch {
		mealName = "lunch"
	} else {
		mealName = "dinner"
	}

	err := cfg.db.View(func(tx *bolt.Tx) error { // nolint: errcheck
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(cfg.userBucket))

		if b == nil {
			return fmt.Errorf("database corrupted: bucket %v not found", cfg.userBucket)
		}

		b.ForEach(func(k, v []byte) error { // nolint: errcheck
			go subscriptionMessage(string(k), menu, subscriptionQR)
			num++
			return nil
		})
		return nil
	})

	if err != nil {
		cfg.debug.Println(err)
	} else {
		cfg.debug.Printf("timed message send attempt for %v message to %v users", mealName, num)
	}
}

func defaultHandler(sender, text string) {
	cfg.debug.Printf("unrecognised command: %v", text)
	responseMessage(sender, unrecognised, defQR)
}

func (e EventHandler) HandleEvent(m []facebook.MessagingEvent) {
	for i := range m {
		r := m[i].Sender.String()
		text := strings.TrimSpace(m[i].Message.Text)
		text = strings.Trim(text, "*_`")
		text = strings.ToLower(text)

		if !strings.HasPrefix(text, e.commandPrefix) {
			defaultHandler(r, text)
			return
		}

		text = strings.TrimPrefix(text, e.commandPrefix)

		switch text {
		case "subscribe", "s":
			subscribeHandler(r)
		case "unsubscribe", "u":
			unsubscribeHandler(r)
		case "help", "h":
			responseMessage(r, helpMessage, helpQR)
		case "lunch", "l":
			menuMessage(r, true)
		case "dinner", "d":
			menuMessage(r, false)
		default:
			defaultHandler(r, text)
		}
	}
}
