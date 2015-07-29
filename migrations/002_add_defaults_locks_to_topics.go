/**
 * In the command line
 *
 * ```sh
 * MONGODB_URL=mongodb://localhost:27017/notifications_testing go run 002_add_defaults_locks_to_topics.go
 * ```
 *
 * `MONGODB_URL` env varilable is not needed if you are running on local, but
 * this can be used to connect to different environments
 */

/**
 * What does this do?
 *
 * - Remove `app_name` property from `topics` collection
 * - Remove previously added `value` property to `topics` collection
 * - Modify `channels` property in `topics` collection from array of strings
 *   to array of objects
 */

package main

import (
	"fmt"
	"log"
	"os"

	models "github.com/bulletind/khabar/db"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	availableCollection = "topics_available"
	topicsCollection    = "topics"
)

type Topic struct {
	models.BaseModel `bson:",inline"`
	User             string   `json:"user" bson:"user"`
	Organization     string   `json:"org" bson:"org"`
	Channels         []string `json:"channels" bson:"channels" required:"true"`
	Ident            string   `json:"ident" bson:"ident" required:"true"`
}

/**
 * Main
 */

func main() {
	session, db := Connect()
	RemoveAppName(db)
	RemoveValue(db)
	ModifyChannels(db)
	session.Close()
	fmt.Println("\n", "Closing mongodb connection")
}

/**
 * Connect to mongo
 */

func Connect() (*mgo.Session, *mgo.Database) {
	uri := os.Getenv("MONGODB_URL")

	if uri == "" {
		uri = "mongodb://localhost:27017/notifications_testing"
	}

	mInfo, err := mgo.ParseURL(uri)
	session, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	session.SetSafe(&mgo.Safe{})
	fmt.Println("Connected to", uri, "\n")

	sess := session.Clone()

	return session, sess.DB(mInfo.Database)
}

/**
 * Remove `app_name` from `topics` collection
 */

func RemoveAppName(db *mgo.Database) (err error) {
	Topics := db.C(topicsCollection)

	change, err := Topics.UpdateAll(
		bson.M{},
		bson.M{
			"$unset": bson.M{
				"app_name": "",
			},
		},
	)
	handle_errors(err)
	fmt.Println("Updated", change.Updated, "documents in `", topicsCollection, "` collection")
	return
}

/**
 * Remove `value` from `topics` collection
 */

func RemoveValue(db *mgo.Database) (err error) {
	Topics := db.C(topicsCollection)

	change, err := Topics.UpdateAll(
		bson.M{},
		bson.M{
			"$unset": bson.M{
				"value": "",
			},
		},
	)
	handle_errors(err)
	fmt.Println("Updated", change.Updated, "documents in `", topicsCollection, "` collection")
	return
}

/**
 * Modify channels array in `topics` collection
 */

func ModifyChannels(db *mgo.Database) (err error) {
	Topics := db.C(topicsCollection)
	notNull := bson.M{"$ne": ""}
	query := bson.M{
		"user": notNull,
		"org":  notNull,
	}
	var topics []Topic

	// Modify user setting

	err = Topics.Find(query).All(&topics)
	handle_errors(err)

	for _, topic := range topics {
		channels := make([]models.Channel, 0)

		for _, name := range topic.Channels {
			channel := new(models.Channel)
			channel.Name = name
			channel.Enabled = true
			channels = append(channels, *channel)
		}

		Topics.Update(bson.M{"_id": topic.Id}, bson.M{
			"$set": bson.M{
				"channels": channels,
			},
		})
	}

	// fmt.Println("Updated", change.Updated, "documents in `", topicsCollection, "` collection")

	// Modify org setting

	query["user"] = ""
	change, err := Topics.UpdateAll(
		query,
		bson.M{
			"$set": map[string][]models.Channel{
				"channels": make([]models.Channel, 0),
			},
		},
	)
	handle_errors(err)
	fmt.Println("Updated", change.Updated, "documents in `", topicsCollection, "` collection")

	return
}

/**
 * Handle errors
 */

func handle_errors(err error) {
	if err != nil {
		log.Printf("Error %v\n", err)
		os.Exit(1)
	}
}
