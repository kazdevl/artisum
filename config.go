package artisum

import (
	"encoding/json"
	"os"
)

type Config struct {
	Urls []string       `json:"urls"`
	Tags []*InterestTag `json:"tags"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c *Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return c, nil
}
