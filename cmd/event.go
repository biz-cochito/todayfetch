package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	_ "modernc.org/sqlite"
)

const (
	BASEURL = "https://api.wikimedia.org/feed/v1/wikipedia/en/onthisday/events/"
	usragt  = "Mozilla/5.0 (X11; Linux x86_64; rv:150.0) Gecko/20100101 Firefox/150.0"
	dbFile  = "todayfetch.db"
)

type WikimediaResponse struct {
	Events []struct {
		Text  string `json:"text"`
		Pages []struct {
			ContentUrls struct {
				Desktop struct {
					Page string `json:"page"`
				} `json:"desktop"`
			} `json:"content_urls"`
		} `json:"pages"`
		Year int `json:"year"`
	} `json:"events"`
}

type CachedEvent struct {
	Year int
	Text string
	URL  string
}

func events() {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}

	date := getDate()
	cached, err := getCachedEvents(db, date)
	if err == nil && len(cached) > 0 {
		fmt.Println("--- Loaded from Cache ---")
		displayCachedEvents(cached, 3)
		return
	}

	fmt.Println("--- Fetching from API ---")
	body := getWikimediaResponse(date)
	var res WikimediaResponse
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		return
	}

	eventsToCache := []CachedEvent{}
	for _, e := range res.Events {
		url := ""
		if len(e.Pages) > 0 {
			url = e.Pages[0].ContentUrls.Desktop.Page
		}
		eventsToCache = append(eventsToCache, CachedEvent{
			Year: e.Year,
			Text: e.Text,
			URL:  url,
		})
	}

	if err := saveEvents(db, date, eventsToCache); err != nil {
		fmt.Printf("Error saving to database: %v\n", err)
	}

	displayCachedEvents(eventsToCache, 3)
}

func initDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		year INTEGER,
		text TEXT,
		url TEXT
	);`
	_, err := db.Exec(query)
	return err
}

func getCachedEvents(db *sql.DB, date string) ([]CachedEvent, error) {
	rows, err := db.Query("SELECT year, text, url FROM events WHERE date = ?", date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []CachedEvent
	for rows.Next() {
		var e CachedEvent
		if err := rows.Scan(&e.Year, &e.Text, &e.URL); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func saveEvents(db *sql.DB, date string, events []CachedEvent) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO events(date, year, text, url) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.Exec(date, e.Year, e.Text, e.URL)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func displayCachedEvents(events []CachedEvent, limit int) {
	if len(events) > limit {
		events = events[:limit]
	}
	for _, e := range events {
		fmt.Printf("[%d] %s\nURL: %s\n\n", e.Year, e.Text, e.URL)
	}
}

func getWikimediaResponse(date string) []byte {
	client := &http.Client{}
	wikimediaurl := constructEventURL(BASEURL, date)

	req, err := http.NewRequest("GET", wikimediaurl, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
	}
	req.Header.Set("User-Agent", usragt)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching URL: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
	}
	return body
}

func constructEventURL(link, date string) string {
	wikimediaurl, err := url.JoinPath(link, date)
	if err != nil {
		return "Failed to construct URL: " + err.Error()
	}
	return wikimediaurl
}

func getDate() string {
	now := time.Now()
	return fmt.Sprintf("%02d/%02d", now.Month(), now.Day())
}
