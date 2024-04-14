package artisum

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type Feeder struct {
	parser   *gofeed.Parser
	feedUrls []string
	from     time.Time
	to       time.Time
}

type Article struct {
	Title       string
	Url         string
	Description string
	Content     string
	Datetime    time.Time
}

func NewFeeder(feedUrls []string, from, to time.Time) *Feeder {
	return &Feeder{
		parser:   gofeed.NewParser(),
		feedUrls: feedUrls,
		from:     from,
		to:       to,
	}
}

func (e *Feeder) ToArticlesMap() (map[string][]*Article, error) {
	var articlesMap = make(map[string][]*Article)
	for _, feedUrl := range e.feedUrls {
		feed, err := e.parser.ParseURL(feedUrl)
		if err != nil {
			return nil, err
		}
		if feed == nil {
			continue
		}

		var articles []*Article
		for _, item := range feed.Items {
			if !e.isWithinRange(item) {
				continue
			}

			articles = append(articles, &Article{
				Title:       item.Title,
				Url:         item.Link,
				Description: item.Description,
				Content:     item.Content,
				Datetime:    *item.PublishedParsed,
			})
		}
		articlesMap[feedUrl] = articles
	}

	return articlesMap, nil
}

func (e *Feeder) isWithinRange(item *gofeed.Item) bool {
	publishDatetime := item.PublishedParsed
	if publishDatetime == nil {
		return false
	}

	if !e.from.IsZero() && publishDatetime.Before(e.from) {
		return false
	}
	if !e.to.IsZero() && publishDatetime.After(e.to) {
		return false
	}

	return true
}
