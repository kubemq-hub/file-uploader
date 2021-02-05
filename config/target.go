package config

import (
	"fmt"
)

var defaultTargetConfig = &Target{
	Address:   "localhost:50000",
	Queue:     "",
	ClientId:  "",
	AuthToken: "",
}

type Target struct {
	Address   string `json:"address"`
	Queue     string `json:"queue"`
	ClientId  string `json:"clientId"`
	AuthToken string `json:"authToken"`
}

func (t *Target) Validate() error {
	if t.Address == "" {
		return fmt.Errorf("kubemq address must have a value")
	}
	if t.Queue == "" {
		return fmt.Errorf("kubemq queue must have a value")
	}
	return nil
}

func (t *Target) Print() {
	log.Infof("Target.Address-> %s", t.Address)
	log.Infof("Target.Queue-> %s", t.Queue)
	log.Infof("Target.ClientId-> %s", t.ClientId)
	log.Infof("Target.AuthToken-> %s", t.AuthToken)
}
