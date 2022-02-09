package poll

import (
	egnyte "github.com/Michael-vdL/egnyte-go-sdk/client"
	es "github.com/elastic/go-elasticsearch/v7"
)

type Poller struct {
	egClient   *egnyte.EventClient
	userClient *egnyte.UserClient
	esClient   es.Client
	state      PollerState
}
