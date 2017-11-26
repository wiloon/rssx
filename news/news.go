package news

import (
	"log"

	"github.com/coreos/bbolt"
	"time"
	"fmt"
	"encoding/json"
)

func init() {

}

var bucket = "NewsBucket"

type Site struct {
	Title string
	News []News
}
type News struct {
	Title string
}

func (site *Site) Append(title string) {
	site.News = append(site.News, News{Title: title})
}
func (site *Site) Save() {
	db, err := bolt.Open("db/boltX.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		log.Println(b)
		return nil
	})

	json, _ := json.Marshal(site.News)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte("oschina"), []byte(json))
		return err
	})

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte("oschina"))
		fmt.Printf("The answer is: %s\n", v)
		return nil
	})

}
func main() {
	db, err := bolt.Open("db/boltX.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		log.Println(b)
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		err := b.Put([]byte("answer"), []byte("42"))
		return err
	})

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		v := b.Get([]byte("answer"))
		fmt.Printf("The answer is: %s\n", v)
		return nil
	})
}
