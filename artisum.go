package artisum

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	configFilePath = ".artisum.json"
)

type Artisum struct {
	feeder          *Feeder
	extracter       *Extracter
	formatter       *ArticleFormatter
	notionRepo      *NotionRepository
	fileRepo        *FileRepository
	now             time.Time
	lastExecuteTime time.Time
}

type SummaryArticle struct {
	Origin   *InterestArticle
	Contents []*FormatContent
}

func NewArtisum(
	numOfSummary int,
	modelName string,
	now time.Time,
	notionRepo *NotionRepository,
	fileRepo *FileRepository,
) (*Artisum, error) {
	conf, err := LoadConfig(configFilePath)
	if err != nil {
		return nil, err
	}

	lastExecuteTime, err := fileRepo.GetLatestExecuteTime()
	if err != nil {
		return nil, err
	}

	if lastExecuteTime.IsZero() {
		lastExecuteTime = now.AddDate(0, 0, -3)
	}

	feeder := NewFeeder(conf.Urls, lastExecuteTime, now)

	extracter, err := NewExtracter(modelName, numOfSummary, conf.Tags)
	if err != nil {
		return nil, err
	}

	formatter, err := NewArticleFormatter(modelName)
	if err != nil {
		return nil, err
	}

	return &Artisum{
		feeder:          feeder,
		extracter:       extracter,
		formatter:       formatter,
		notionRepo:      notionRepo,
		fileRepo:        fileRepo,
		now:             now,
		lastExecuteTime: lastExecuteTime,
	}, nil
}

func (a *Artisum) Summary(ctx context.Context) error {
	slog.Info("start summary", slog.String("lastExecuteTime", a.lastExecuteTime.Format(time.DateOnly)), slog.String("now", a.now.Format(time.DateOnly)))
	if a.isExecutedToday() {
		slog.Info("already executed today")
		return nil
	}

	slog.Info("start collect articles...")
	feedArticleMap, err := a.feeder.ToArticlesMap()
	if err != nil {
		return err
	}
	slog.Info("collected")
	if len(feedArticleMap) == 0 {
		slog.Info("no articles from feeds")
		return nil
	}

	slog.Info("extracting...")
	articles, err := a.extracter.Extract(ctx, feedArticleMap)
	if err != nil {
		return err
	}
	slog.Info("extracted")

	var eg errgroup.Group
	for _, article := range articles {
		eg.Go(func() error {
			slog.Info("formatting...", slog.String("url", article.URL), slog.String("tag", article.Tag))
			FormatContents, err := a.formatter.Format(ctx, article.URL)
			if err != nil {
				return err
			}

			slog.Info("formatted")
			return a.notionRepo.SaveSummaryResult(ctx, &SummaryArticle{
				Origin:   article,
				Contents: FormatContents,
			})
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return a.fileRepo.SaveExecuteTime(a.now)
}

func (a *Artisum) isExecutedToday() bool {
	return a.now.Format(time.DateOnly) == a.lastExecuteTime.Format(time.DateOnly)
}
