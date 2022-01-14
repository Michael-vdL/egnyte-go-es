package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/Michael-vdL/egnyte-go-es/poll"
	egnyte "github.com/Michael-vdL/egnyte-go-sdk/client"
	elastic "github.com/elastic/go-elasticsearch/v7"
)

var (
	egnyteConfigFile  = flag.String("egnyte-cfg", "./config/egnyte.json", "Relative path to egnyte configuration file")
	elasticConfigFile = flag.String("elastic-cfg", "./config/elastic.json", "Relative path to elastic configuration file")
  pollingInterval = flag.Int("pollingInterval", 5, "Interval to poll events from egnyte in minutes")
  historyLimit = flag.Int("historyLimit", 1000, "How far back the initial poll of Egnyte will go. Default is 1000. Will reference state file and current cursor")
)


func main() {
  flag.Parse()

  egnyteConfig, err := egnyte.ConfigurationFromFile(*egnyteConfigFile)
  if err != nil {
    log.Fatal(err)
  }

  egnyteClient := egnyte.NewClient(egnyteConfig.SubDomain, egnyteConfig.APIVersion)
  err = egnyteClient.Authenticate(egnyteConfig.Username, egnyteConfig.Password, egnyteConfig.APIKey, egnyteConfig.APISecret)
  if err != nil {
    log.Fatal(err)
  }
  egEventClient := &egnyte.EventClient{Client: *egnyteClient}

  esClient, _ := elastic.NewDefaultClient() // TOOD: Implement ES Client from Config File
  poller := poll.New(egEventClient, *esClient)

  poller.Poll(context.Background(), time.Duration(*pollingInterval) * time.Minute, *historyLimit)
}