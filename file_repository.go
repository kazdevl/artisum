package artisum

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type FileRepository struct {
	summaryPath     string
	feedPath        string
	executeTimePath string
}

func NewFileRepository(dirPath, modelName string, now time.Time) *FileRepository {
	nowStr := now.Format(time.DateOnly)
	summaryPath := fmt.Sprintf("%s/%s_%s_summary", dirPath, modelName, nowStr)
	feedPath := fmt.Sprintf("%s/%s_%s_feed.json", dirPath, modelName, nowStr)
	executeTimePath := fmt.Sprintf("%s/execute_time", dirPath)
	return &FileRepository{
		summaryPath:     summaryPath,
		feedPath:        feedPath,
		executeTimePath: executeTimePath,
	}
}

func (f *FileRepository) SaveExecuteTime(t time.Time) error {
	file, err := os.Create(f.executeTimePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(t.Format(time.DateOnly)); err != nil {
		return err
	}

	return nil
}

func (f *FileRepository) GetLatestExecuteTime() (time.Time, error) {
	file, err := os.Open(f.executeTimePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	v := scanner.Text()
	t, err := time.Parse(time.DateOnly, scanner.Text())
	if err != nil {
		return time.Time{}, err
	}

	slog.Info("latest execute time", slog.String("time", v), slog.Time("parsed", t))

	return t, nil
}
