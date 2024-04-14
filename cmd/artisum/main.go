package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/kazdevl/artisum"
)

var (
	numOfSummaryF int
	modelNameF    string
)

const (
	outputDirPath = ".artisum/result"
)

var (
	notionToken      = os.Getenv("NOTION_TOKEN")
	notionDatabaseID = os.Getenv("NOTION_DATABASE_ID")
)

func init() {
	if _, err := os.Stat(outputDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDirPath, 0755); err != nil {
			panic(err)
		}
	}
}

func main() {
	flag.IntVar(&numOfSummaryF, "num", 3, "number of summary")
	flag.StringVar(&modelNameF, "model", "gpt-4-turbo", "model name")
	flag.Parse()

	slog.Info("start artisum", slog.String("model", modelNameF), slog.Int("num", numOfSummaryF))

	if err := run(); err != nil {
		panic(err)
	}

	slog.Info("end artisum")
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	now := time.Now()
	notionRepo := artisum.NewNotionRepository(notionToken, notionDatabaseID)
	fileRepo := artisum.NewFileRepository(outputDirPath, modelNameF, time.Now())
	a, err := artisum.NewArtisum(numOfSummaryF, modelNameF, now, notionRepo, fileRepo)
	if err != nil {
		return err
	}

	return a.Summary(ctx)
}
