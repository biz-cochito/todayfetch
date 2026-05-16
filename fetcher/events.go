package fetcher

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/savioxavier/termlink"
)

const (
	BASEURL = "https://api.wikimedia.org/feed/v1/wikipedia/en/onthisday/events/"
	usragt  = "Mozilla/5.0 (X11; Linux x86_64; rv:150.0) Gecko/20100101 Firefox/150.0"
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

func FetchEvents(db *sql.DB, limit int) {
	date := GetDate()
	cached, err := getCachedEvents(db, date)
	if err == nil && len(cached) > 0 {
		fmt.Println("--- History (from Cache) ---")
		displayEvents(cached, limit)
		return
	}

	fmt.Println("--- History (Fetching...) ---")
	body := getWikimediaResponse(date)
	var res WikimediaResponse
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	eventsToCache := []CachedEvent{}
	for _, e := range res.Events {
		link := ""
		if len(e.Pages) > 0 {
			link = e.Pages[0].ContentUrls.Desktop.Page
		}
		eventsToCache = append(eventsToCache, CachedEvent{
			Year: e.Year,
			Text: e.Text,
			URL:  link,
		})
	}

	if err := saveEvents(db, date, eventsToCache); err != nil {
		fmt.Printf("Error saving cache: %v\n", err)
	}

	displayEvents(eventsToCache, limit)
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

func displayEvents(events []CachedEvent, limit int) {
	if len(events) > limit {
		events = events[:limit]
	}
	for _, e := range events {
		fmt.Printf(`[%d] %s`, e.Year, termlink.ColorLink(e.Text, e.URL, "green"))
		// fmt.Printf("[%d] %s\n", e.Year, e.Text)
		// if e.URL != "" {
		// 	fmt.Pr("      URL: %s\n", e.URL)
		// }
		// fmt.Println()
	}
}

func getWikimediaResponse(date string) []byte {
	client := &http.Client{}
	u, _ := url.JoinPath(BASEURL, date)

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", usragt)
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}

func GetDate() string {
	now := time.Now()
	return fmt.Sprintf("%02d/%02d", now.Month(), now.Day())
}
