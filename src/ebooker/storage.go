package ebooker

import (
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"strconv"
	"sort"
)

type DataHandle struct {
    handle *sql.DB
    logger *LogMaster
}


// Ensures we've got a valid instance of the database, and if not, creates one.
func GetDataHandle(filename string, logger *LogMaster) DataHandle {

	db, err := sql.Open("sqlite3", filename)
    handle := DataHandle{db, logger}
	if err != nil {
	    logger.StatusWrite("sql.Open returned non-nil error!\n")
	    logger.DebugWrite("sql.Open returned error: %v\n", err)
	    return handle
	}

	sqls := []string{ "CREATE TABLE Tweets (Id TEXT NOT NULL, Screen_Name TEXT NOT NULL, Content TEXT NOT NULL)" }
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil && err.Error() != "table Tweets already exists" {
	        logger.StatusWrite("sql.Open returned unexpected error on DataHandle Aquisition.\n")
	        logger.DebugWrite("Error: %v\n", err)
			return handle
		}
	}

	return handle
}


// Retrieves all tweets we have for a given user.
func (dh DataHandle) GetTweetsFromStorage(username string) Tweets {
    db := dh.handle

    queryStr := "SELECT Id, Content FROM Tweets WHERE Screen_name = ?"
    dh.logger.DebugWrite("Query on datastore is %s on %s\n", queryStr, username)

	rows, err := db.Query(queryStr, username)
	if err != nil {
		dh.logger.StatusWrite("Unexpected error on query to datastore\n")
		dh.logger.DebugWrite("Error was %v\n", queryStr, username, err)
		return Tweets{}
	}
	defer rows.Close()

	var oldtweets Tweets
	for rows.Next() {
		var id, text string
		rows.Scan(&id, &text)
		idInt, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
            dh.logger.StatusWrite("Tweet id %s not able to form valid ID in ParseUint\n", id)
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
		dh.logger.StatusWrite("Unexpected Error in Aquiring a Transaction to Insert into.\n")
		dh.logger.DebugWrite("Error is %v\n", err)
		return
	}

    insertStr := "INSERT INTO Tweets (Id, Screen_name, Content) VALUES (?, ?, ?)"
    dh.logger.DebugWrite("SQL Statement to insert is \"%s\"\n", insertStr)

	stmt, err := tx.Prepare(insertStr)
	if err != nil {
		dh.logger.StatusWrite("Unexpected Error in Preparing INSERT Statement.\n")
		dh.logger.DebugWrite("Error is %v\n", err)
		return
	}
	defer stmt.Close()

    for _, tweet := range newTweets {
        idStr := strconv.FormatUint(tweet.Id, 10)
        _, err = stmt.Exec(idStr, username, tweet.Text)
        if err != nil {
            dh.logger.StatusWrite("Unexpected Error in Executing INSERT Statement.\n")
            dh.logger.DebugWrite("Error is %v\n", err)
            return
        }
    }
	tx.Commit()
}


// should defer on DataHandles!
func (dh DataHandle) Cleanup() {
	dh.handle.Close()
}
