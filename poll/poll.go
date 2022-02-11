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

	// * Right now, it is too easy to overwhelm the poller with events. moving a folder higher up in the tree could easily generate thousands of events.
	// * Need to identify a good way to handle surges of events. Something conditional that would maybe filter dynamically if
	for {
		log.Default().Println("Polling Again")
		events, err := p.egClient.GetEvents(strconv.Itoa(p.state.LastCursor))
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
			p.state.LastCursor = events.LatestId
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
// TODO: Move this to a middleware function so data can be cleaned up. For now need to leverage some other data store.
func (p *Poller) postEgnyteEvents(events *egnyte.Events) {
	log.Default().Println("Poller: Starting - posting events")
	// Enrich data with usernames before sending to Egnyte...
	actorMap := make(map[int]string)
	for _, user := range p.state.Users.Resources {
		actorMap[user.ID] = user.UserName
	}
	for _, event := range events.Events {
		if _, ok := actorMap[event.Actor]; !ok {
			// Need to find a way to stagger this call and optimize... Rate limits are going to kill me
			newUser, err := p.userClient.GetUserById(strconv.Itoa(event.Actor))
			if err != nil {
				log.Fatal(err)
			}
			p.state.Users.Resources = append(p.state.Users.Resources, *newUser)
			actorMap[newUser.ID] = newUser.UserName
		}
		res, err := p.esClient.Index(
			"egnyte-events-",
			esutil.NewJSONReader(event),
		)
		defer res.Body.Close() // Why can't I defer this? (Originally in Loop)
		if err != nil {
			log.Panic(err)
		}
	}
	log.Default().Println("Poller: Complete - posting events")
	p.updateStateFile()
}

func New(egClient *egnyte.EventClient, userClient *egnyte.UserClient, esClient es.Client) *Poller {

	// Initialize Poller object
	poller := &Poller{
		egClient:   egClient,
		userClient: userClient,
		esClient:   esClient,
	}

	// Generate Poller State
	poller.getInitialState()

	return poller
}
