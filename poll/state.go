package poll

import (
	"encoding/gob"
	"errors"
	"log"
	"os"

	egnyte "github.com/Michael-vdL/egnyte-go-sdk/client"
)

// TODO Move State functions to state struct instead of poller struct

type PollerState struct {
	Users        egnyte.Users
	LastCursor   int
	LatestCursor int
}

func (p *Poller) setInitialUsers() {
	users, err := p.userClient.GetUsers()
	if err != nil {
		log.Fatal(err)
	}
	p.state.Users = *users
}


func (p *Poller) setInitialPollingCursor(historyLimit int) {
	cursor, err := p.egClient.GetCursor()
	if err != nil {
		log.Panic(err)
	}

	p.state.LatestCursor = cursor.LatestEventId
	if p.state.LastCursor+historyLimit >= cursor.LatestEventId {
		p.state.LastCursor = cursor.OldestEventId
	} else {
		p.state.LastCursor = cursor.LatestEventId - historyLimit
	}
}


func (p *Poller) getInitialState() {
	if _, err := os.Stat("service.state"); err == nil {
		file, err := os.Open("service.state")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		decoder := gob.NewDecoder(file)
		pollerState := PollerState{}
		decoder.Decode(&pollerState)
		p.state = pollerState
	} else {
		p.setInitialUsers()
		p.setInitialPollingCursor(1000) // TODO: Move this to an options struct for poller
	}
}


func (p *Poller) updateStateFile() {
	  // Check if state file exists or if it needs to be created
		if _, err := os.Stat("service.state"); err == nil { // If file already exists, update
			file, err := os.OpenFile("service.state", os.O_WRONLY, 0775)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			encoder := gob.NewEncoder(file)
			err = encoder.Encode(p.state)
			if err != nil {
				log.Fatal(err)
			}
		} else if errors.Is(err, os.ErrNotExist) { // If file does not exist, write file to state
			file, _ := os.Create("service.state")
  		defer file.Close()
  		encoder := gob.NewEncoder(file)
  		err = encoder.Encode(p.state)
  		if err != nil {
  		  log.Fatal(err)
  		}
		} else {
			log.Fatal(err)
		}

    log.Default().Println(p.state)
}