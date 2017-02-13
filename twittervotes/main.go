package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	nsq "github.com/bitly/go-nsq"

	mgo "gopkg.in/mgo.v2"
)

func main() {
	var stoplock sync.Mutex // this protects stop
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if err := dialdb(); err != nil {
		log.Fatalln("failed to dial MongoDB:", err)
	}
	defer closedb()

	// kicking things off
	votes := make(chan string) // chan for votes
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}

var (
	db     *mgo.Session
	dbHost string = "mongodb://" + os.Getenv("SP_MONGO_HOST") + "/test"
)

// dialdb opens connection to mongo dbHost
func dialdb() error {
	// enablling debug
	// mgo.SetDebug(true)

	// var aLogger *log.Logger
	// aLogger = log.New(os.Stderr, "", log.LstdFlags)
	// mgo.SetLogger(aLogger)

	var err error
	log.Printf("dialing mongodb: %s", dbHost)
	db, err = mgo.Dial(dbHost)
	return err
}

// closedb closes connection to mongo dbHost
func closedb() {
	db.Close()
	log.Println("close database connection")
}

// poll is kept small as we only care about Options
// the polls document could contain more
type poll struct {
	Options []string
}

// loadOptions loads the polls collection from the ballots database
// using Find(nil) means do not filter
// the Iter method allos us to access each poll one by one
// this is memory efficient because we only ever using a single poll object
// Using the mgo .All method could get out of control depending on its size
func loadOptions() ([]string, error) {
	var options []string
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	var p poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

var (
	nsqHost string = os.Getenv("SP_NSQ_HOST")
	nsqPort string = ":4150"
)

// publishVotes starts a go routine that publishes twitter votes to NSQ
// uses a signal channel when the go routine needs to be closed
func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, err := nsq.NewProducer(nsqHost+nsqPort, nsq.NewConfig())
	if err != nil {
		log.Println("Error creating NSQ Producer:", err)
		stopchan <- struct{}{}
		return stopchan
	}
	go func() {
		for vote := range votes {
			pub.Publish("votes", []byte(vote)) // publish vote
		}
		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
		stopchan <- struct{}{}
	}()
	return stopchan
}
