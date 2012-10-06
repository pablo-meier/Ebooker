package main

/*
Contains functions + data structures for persistent data storage. Currently
using sqlite3.
*/

import (
	_ "github.com/mattn/go-sqlite3"

	"ebooker/logging"
	"ebooker/oauth1"

	"database/sql"
	"sort"
	"strconv"
)

// Top-level object that maintains the database connection.
type DataHandle struct {
	handle *sql.DB
	logger *logging.LogMaster
}

// Cleanup function for resources used by `storage`. Should defer this call
// after obtaining a DataHandle!
func (dh DataHandle) Cleanup() {
	dh.handle.Close()
}

// Ensures we've got a valid instance of the database, and if not, creates one
// with the appropriate tables.
func getDataHandle(filename string, logger *logging.LogMaster) DataHandle {

	db, err := sql.Open("sqlite3", filename)
	handle := DataHandle{db, logger}
	if err != nil {
		logger.StatusWrite("sql.Open returned non-nil error!\n")
		logger.DebugWrite("sql.Open returned error: %v\n", err)
		return handle
	}

	sqls := []string{"CREATE TABLE Tweets (Id TEXT NOT NULL, Screen_Name TEXT NOT NULL, Content TEXT NOT NULL)",
		"CREATE TABLE TwitterUsers (Screen_Name TEXT NOT NULL, Token TEXT NOT NULL, Token_Secret TEXT NOT NULL)"}
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil && err.Error() != "table Tweets already exists" && err.Error() != "table TwitterUsers already exists" {
			logger.StatusWrite("sql.Exec returned unexpected error on DataHandle Aquisition.\n")
			logger.DebugWrite("Error: %v\n", err)
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
		oldtweets = append(oldtweets, TweetData{idInt, text})
	}

	sort.Sort(oldtweets)
	return oldtweets
}

// Inserts tweets into persistent storage.
func (dh DataHandle) InsertFreshTweets(username string, newTweets Tweets) {
	db := dh.handle

	tx, err := db.Begin()
	if err != nil {
		dh.logger.StatusWrite("Unexpected Error in Aquiring a Transaction to Insert into.\n")
		dh.logger.DebugWrite("Error is %v\n", err)
		return
	}

	insertStr := "INSERT INTO Tweets (Id, Screen_name, Content) VALUES (?, ?, ?)"

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

// Retrieves the "oauth_token" and "oauth_token_secret" for a given user, if we
// have it. If we don't, we state so in the second parameter.
func (dh DataHandle) getUserAccessToken(username string) (*oauth1.Token, bool) {
	db := dh.handle

	queryStr := "SELECT Token, Token_Secret FROM TwitterUsers WHERE Screen_name = ?"

	rows, err := db.Query(queryStr, username)
	if err != nil {
		dh.logger.StatusWrite("Unexpected error on query to datastore\n")
		dh.logger.DebugWrite("Error was %v\n", queryStr, username, err)
		return nil, false
	}
	defer rows.Close()

	var token, tokenSecret string
	length := 0
	for rows.Next() {
		rows.Scan(&token, &tokenSecret)
		length += 1
	}

	if length == 0 {
		return nil, false
	}
	return &oauth1.Token{token, tokenSecret}, true
}

// Inserts tweets into persistent storage.
func (dh DataHandle) insertUserAccessToken(username string, token *oauth1.Token) {
	db := dh.handle

	tx, err := db.Begin()
	if err != nil {
		dh.logger.StatusWrite("Unexpected Error in Aquiring a Transaction to Insert into.\n")
		dh.logger.DebugWrite("Error is %v\n", err)
		return
	}

	insertStr := "INSERT INTO TwitterUsers (Screen_name, Token, Token_Secret) VALUES (?, ?, ?)"

	stmt, err := tx.Prepare(insertStr)
	if err != nil {
		dh.logger.StatusWrite("Unexpected Error in Preparing INSERT Statement.\n")
		dh.logger.DebugWrite("Error is %v\n", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token.OAuthToken, token.OAuthTokenSecret)
	if err != nil {
		dh.logger.StatusWrite("Unexpected Error in Executing INSERT Statement.\n")
		dh.logger.DebugWrite("Error is %v\n", err)
		return
	}
	tx.Commit()
}
