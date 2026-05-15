package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)
const BASEURL = "https://api.wikimedia.org/feed/v1/wikipedia/en/onthisday/events/"
const usragt = "Mozilla/5.0 (X11; Linux x86_64; rv:150.0) Gecko/20100101 Firefox/150.0"

type WikimediaResponse struct {
	Events []struct {
		Text  string `json:"text"`
		Pages []struct {
			Type         string `json:"type"`
			Title        string `json:"title"`
			Displaytitle string `json:"displaytitle"`
			Namespace    struct {
				ID   int    `json:"id"`
				Text string `json:"text"`
			} `json:"namespace"`
			WikibaseItem string `json:"wikibase_item"`
			Titles       struct {
				Canonical  string `json:"canonical"`
				Normalized string `json:"normalized"`
				Display    string `json:"display"`
			} `json:"titles"`
			Pageid    int `json:"pageid"`
			Thumbnail struct {
				Source string `json:"source"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"thumbnail"`
			Originalimage struct {
				Source string `json:"source"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"originalimage"`
			Lang              string    `json:"lang"`
			Dir               string    `json:"dir"`
			Revision          string    `json:"revision"`
			Tid               string    `json:"tid"`
			Timestamp         time.Time `json:"timestamp"`
			Description       string    `json:"description"`
			DescriptionSource string    `json:"description_source"`
			ContentUrls       struct {
				Desktop struct {
					Page      string `json:"page"`
					Revisions string `json:"revisions"`
					Edit      string `json:"edit"`
					Talk      string `json:"talk"`
				} `json:"desktop"`
				Mobile struct {
					Page      string `json:"page"`
					Revisions string `json:"revisions"`
					Edit      string `json:"edit"`
					Talk      string `json:"talk"`
				} `json:"mobile"`
			} `json:"content_urls"`
			Extract         string `json:"extract"`
			ExtractHTML     string `json:"extract_html"`
			Normalizedtitle string `json:"normalizedtitle"`
			Coordinates     struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"coordinates,omitempty"`
		} `json:"pages"`
		Year int `json:"year"`
	} `json:"events"`
}


func events() {
	body := getWikimediaResponse()
	displayEvents(body)
}

func displayEvents(body []byte) {
	var res WikimediaResponse
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		return
	}
	if len(res.Events) > 10 {
		res.Events = res.Events[:10]
	}
	for _, rec := range res.Events {
		fmt.Println(rec.Text)
	}
}

func getWikimediaResponse() []byte {
	client := &http.Client{}
	wikimediaurl := constructEventURL(BASEURL)

	req, err := http.NewRequest("GET", wikimediaurl, nil)
	if err != nil {
		fmt.Printf("Error fetching URL: %v\n", err)
	}
	req.Header.Set("User-Agent", usragt)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching URL: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body) // body is []byte
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
	}
	return body
}

func PrettyPrint(response WikimediaResponse) string {

    s, _ := json.MarshalIndent(response, "", "\t")

    return string(s)

}

func constructEventURL(link string) string {
	wikimediaurl, err := url.JoinPath(link, getDate())
	if err != nil {
		return "Failed to construct URL: " + err.Error()
	}
	return wikimediaurl
}

func getDate() string {
	now := time.Now()
	monthInt := int(now.Month())
	dayInt := int(now.Day())
	month := formatDate(monthInt)
	day := formatDate(dayInt)
	dateStr := month + "/" + day
	fmt.Println(dateStr)
	return dateStr
}

func formatDate(x int) string {
	var str string
	if x < 10 {
		str = "0" + strconv.Itoa(x)
	} else {
		str = strconv.Itoa(x)
	}
	return str
}
