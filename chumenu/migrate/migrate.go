// Migration script from chumenu JSON databases
package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"os"
	"time"
)

type jsonDB map[string]bool

func main() {
	if len(os.Args) != 4 {
		panic("Usage: migrate <json file> <boltdb file> <user bucket>")
	}

	b, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		panic(err)
	}

	db := jsonDB{}
	json.Unmarshal(b, &db)

	bdb, err := bolt.Open(os.Args[2], 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	err = bdb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(os.Args[3]))
		if err != nil {
			return err
		}

		for k := range db {
			err = b.Put([]byte(k), []byte{})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	err = bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(os.Args[3]))

		b.ForEach(func(k, v []byte) error {
			fmt.Println(string(k))
			return nil
		})
		return nil
	})
}
