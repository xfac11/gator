package rssfeed

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "gator")
	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Could not send a request. Error: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode > 300 {
		return nil, fmt.Errorf("Non successful status code: %d.", response.StatusCode)
	}

	decoder := xml.NewDecoder(response.Body)
	if decoder == nil {
		return nil, fmt.Errorf("Ci")
	}

	var feed RSSFeed
	err = decoder.Decode(&feed)
	if err != nil {
		return nil, fmt.Errorf("Could not decode xml to RssFeed. Error: %w", err)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i, _ := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}
