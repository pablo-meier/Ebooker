package ebooker

import (
	_ "github.com/mattn/go-sqlite3"

	"fmt"
	"database/sql"
	"log"
	"strconv"
	"sort"
)

type DataHandle struct {
    handle *sql.DB
}


// Ensures we've got a valid instance of the database, and if not, creates one.
func GetDataHandle(filename string) DataHandle {

	db, err := sql.Open("sqlite3", filename)
    handle := DataHandle{db}
	if err != nil {
		fmt.Println(err)
	    log.Fatal(err)
	    return handle
	}

	sqls := []string{ "CREATE TABLE Tweets (Id TEXT NOT NULL, Screen_Name TEXT NOT NULL, Content TEXT NOT NULL)" }
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil && err.Error() != "table Tweets already exists" {
			fmt.Printf("%q: %s\n", err, sql)
			return handle
		}
	}

	return handle
}


// Retrieves all tweets we have for a given user.
func (dh DataHandle) GetTweetsFromStorage(username string) Tweets {

    db := dh.handle

	rows, err := db.Query("SELECT Id, Content FROM Tweets WHERE Screen_name = ?", username)
	if err != nil {
		fmt.Println(err)
		return Tweets{}
	}
	defer rows.Close()

	var oldtweets Tweets
	for rows.Next() {
		var id, text string
		rows.Scan(&id, &text)
		idInt, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
            log.Fatal(fmt.Sprintf("Tweet id %s not able to form valid ID in ParseUint\n", id))
		}
		oldtweets = append(oldtweets, TweetData{ idInt, text })
	}
	rows.Close()

    sort.Sort(oldtweets)

    return oldtweets
}


// Inserts the tweets into the persistent storage.
func (dh DataHandle) InsertFreshTweets(username string, newTweets Tweets) {

    db := dh.handle

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO Tweets (Id, Screen_name, Content) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()

    for _, tweet := range newTweets {
        idStr := strconv.FormatUint(tweet.Id, 10)
        _, err = stmt.Exec(idStr, username, tweet.Text)
        if err != nil {
            fmt.Println(err)
            return
        }
    }
	tx.Commit()
}


// should defer on DataHandles!
func (dh DataHandle) Cleanup() {
	dh.handle.Close()
}
