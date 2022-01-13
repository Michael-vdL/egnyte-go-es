package poll

import (
	"context"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	egnyte "github.com/Michael-vdL/egnyte-go-sdk/client"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

const DefaultPollingInterval = 1 * time.Minute

type Poller struct {
  egClient *egnyte.EventClient
  esClient es.Client
  lastCursor int
  latestCursor int
}

// For now, just assuming poller is only for events
func (p *Poller) Poll(ctx context.Context, interval time.Duration) {
  // TODO: implement reusable polling - idea would be to have a poll method on clients that can be polled.
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

  eventCursor, err := p.egClient.GetCursor()
  if err != nil {
    log.Panic(err)
  }
  p.lastCursor = eventCursor.OldestEventId
  p.latestCursor = eventCursor.LatestEventId

  log.Default().Println("Polling on a pole")
  for range t.C {
    for p.lastCursor < p.latestCursor {
      events, err := p.egClient.GetEvents(strconv.Itoa(p.lastCursor))
      if err != nil {
        if err.Error() == "get events failed. status code: 403" {
          log.Default().Println("Poller: Hit Egnyte Rate Limit, sleeping for interval")
          t.Reset(interval)
          break
        } else {
          log.Panic(err)
        }
      }
      p.lastCursor = events.LatestId
      for _, event := range events.Events {
        log.Println(event)
        res, err := p.esClient.Index(
          "egnyte-events-",
          esutil.NewJSONReader(event),
        )
  
        body, _ := ioutil.ReadAll(res.Body)
        res.Body.Close() // Why can't I defer this?
        log.Println(string(body))
  
        if err != nil {
          log.Panic(err)
        }
      }  

    }    
    select {
    case <-ctx.Done():
      log.Default().Panicln("stopping poller")
    case <-t.C:
    }
    
  }
}

func New(egClient *egnyte.EventClient, esClient es.Client) *Poller {
  return &Poller{
    egClient: egClient,
    esClient: esClient,
  }
}