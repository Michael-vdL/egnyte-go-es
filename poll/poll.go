package poll

import (
	"context"
	"log"
	"strconv"
	"time"

	egnyte "github.com/Michael-vdL/egnyte-go-sdk/client"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

const DefaultPollingInterval = 5 * time.Minute

type Poller struct {
	egClient     *egnyte.EventClient
	esClient     es.Client
	lastCursor   int
	latestCursor int
}

// For now, just assuming poller is only for events
func (p *Poller) Poll(ctx context.Context, interval time.Duration, historyLimit int) {
	// TODO: implement reusable polling - idea would be to have a pollable interface for variables endpoints that can be polled.
	// if len(endpoints) == 0 {
	//   log.Default().Println("Could not initiate poller: No endpoints to poll")
	//   return
	// }

	if interval <= 0 {
		log.Default().Println("Invalid interval, continuing with default of 5m")
		interval = DefaultPollingInterval
	}

	t := time.NewTicker(interval)
	defer t.Stop()

	p.SetInitialPollingCursor(historyLimit) 
 
  // * Right now, it is too easy to overwhelm the poller with events. moving a folder higher up in the tree could easily generate thousands of events.
  // * Need to identify a good way to handle surges of events. Something conditional that would maybe filter dynamically if 
	for {
		log.Default().Println("Polling Again")
		events, err := p.egClient.GetEvents(strconv.Itoa(p.lastCursor))
		if err != nil {
			if err.Error() == "get events failed. status code: 403" {
				log.Default().Println("Poller: Hit Egnyte Rate Limit, sleeping for interval")
				t.Reset(interval)
				break
			}
			if err.Error() == "get events failed. status code: 204" { // TODO: Handle this in SDK
				log.Default().Println("Poller: No more events to poll")
			} else {
				log.Panic(err)
			}
		}
    if events != nil {
      p.lastCursor = events.LatestId
      p.postEgnyteEvents(events)
    }
		select {
		case <-ctx.Done():
			log.Default().Panicln("stopping poller")
		case <-t.C:
		}

	}
}

// Temperary... need to make this a generic function for any pollable endpoint (if possible)
func (p *Poller) postEgnyteEvents(events *egnyte.Events) {
	log.Default().Println("Poller: Starting - posting events")
	for _, event := range events.Events {
		res, err := p.esClient.Index(
			"egnyte-events-",
			esutil.NewJSONReader(event),
		)
		res.Body.Close() // Why can't I defer this? (Originally in Loop)
		if err != nil {
			log.Panic(err)
		}
	}
	log.Default().Println("Poller: Complete - posting events")
}

func (p *Poller) SetInitialPollingCursor(historyLimit int) {
	cursor, err := p.egClient.GetCursor()
	if err != nil {
		log.Panic(err)
	}

	p.latestCursor = cursor.LatestEventId
	if p.lastCursor+historyLimit >= cursor.LatestEventId {
		p.lastCursor = cursor.OldestEventId
	} else {
		p.lastCursor = cursor.LatestEventId - historyLimit
	}
}

func New(egClient *egnyte.EventClient, esClient es.Client) *Poller {
	return &Poller{
		egClient: egClient,
		esClient: esClient,
	}
}
