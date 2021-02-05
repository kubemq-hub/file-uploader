package config

import (
	"fmt"
)

var defaultSourceConfig = &Source{
	Root:                "./",
	ConcurrentSenders:   1,
	PollIntervalSeconds: 60,
	BucketType:          "aws",
	BucketName:          "bucket",
}

type Source struct {
	Root                string `json:"root"`
	ConcurrentSenders   int    `json:"concurrent_senders"`
	PollIntervalSeconds int    `json:"pollIntervalSeconds"`
	BucketType          string `json:"bucketType"`
	BucketName          string `json:"bucketName"`
}

func (s *Source) Validate() error {
	if s.Root == "" {
		return fmt.Errorf("root path cannot be empty")
	}
	if s.ConcurrentSenders <= 0 {
		return fmt.Errorf("concurrent senders must be > 0")
	}
	if s.PollIntervalSeconds <= 0 {
		return fmt.Errorf("poll intreval must be > 0 secnods")
	}
	switch s.BucketType {
	case "aws", "gcp", "minio":
	default:
		return fmt.Errorf("invalid bucket type")
	}
	if s.BucketName == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}
	return nil
}
func (s *Source) Print() {
	log.Infof("Source.Root-> %s", s.Root)
	log.Infof("Source.ConcurrentSenders-> %d", s.ConcurrentSenders)
	log.Infof("Source.PollIntervalSeconds-> %d", s.PollIntervalSeconds)
	log.Infof("Source.BucketType-> %s", s.BucketType)
	log.Infof("Source.BucketName-> %s", s.BucketName)
}
