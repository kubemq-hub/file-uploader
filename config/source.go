package config

import (
	"fmt"
	"strings"
)

var defaultSourceConfig = &Source{
	SyncPath:            "./",
	Include:             nil,
	Ignore:              nil,
	PollIntervalSeconds: 60,
}

type Source struct {
	SyncPath            string   `json:"syncPath"`
	Include             []string `json:"include"`
	Ignore              []string `json:"ignore"`
	PollIntervalSeconds int      `json:"pollIntervalSeconds"`
}

func (s *Source) Validate() error {
	if s.SyncPath == "" {
		return fmt.Errorf("sync path cannot be empty")
	}
	if s.PollIntervalSeconds <= 0 {
		return fmt.Errorf("poll intreval must be > 0 secnods")
	}
	return nil
}
func (s *Source) Print() {
	log.Infof("Source.SyncPath-> %s", s.SyncPath)
	log.Infof("Source.Include-> %s", strings.Join(s.Include, ","))
	log.Infof("Source.Ignore-> %s", strings.Join(s.Ignore, ","))
	log.Infof("Source.PollIntervalSeconds-> %d", s.PollIntervalSeconds)
}
