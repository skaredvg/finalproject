package api

import (
	"encoding/xml"
	"io"
	"net/http"
)

// Объект-публикация
type RSSPublication struct {
	Title       string `xml:"title"`
	Description string `xml:"description,omitempty"`
	PubTime     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

// Объект-массив публикаций
type RSSChannel struct {
	Title        string
	Publications []RSSPublication `xml:"item"`
}

// Объект - RSS-рассылка
type RSSNewsFeed struct {
	Site    string
	XMLName xml.Name
	Channel RSSChannel `xml:"channel"`
}

// Конструктор объекта RSS-ленты
func NewRSSNewsFeed(l string) *RSSNewsFeed {
	return &RSSNewsFeed{Site: l}
}

// Распарсить RSS-ленту по ссылке
func (rnf *RSSNewsFeed) ProcessLink() error {
	resp, err := http.Get(rnf.Site)
	if err != nil {
		return err
	}

	b, _ := io.ReadAll(resp.Body)
	if err := xml.Unmarshal(b, &rnf); err != nil {
		return err
	}
	return nil
}
