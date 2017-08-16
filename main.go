package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	flag "github.com/jessevdk/go-flags"
	"github.com/maplebed/libplum"
	"github.com/maplebed/plumd/actions"
)

type Options struct {
	Email     string `short:"e" long:"email" descrption:"Email address to authenticate with the Plum Web API"`
	Password  string `short:"p" long:"password" descrption:"Password to authenticate with the Plum Web API"`
	StateFile string `short:"f" long:"file" description:"location to store state file" default:"/var/lib/plumd.state"`
	Debug     bool   `short:"d" long:"debug" description:"enable debugging output"`
	Test      bool   `short:"t" long:"test" description:"run the app with mocked plum"`
}

func main() {
	var options Options
	flagParser := flag.NewParser(&options, flag.Default)
	flagParser.Parse()
	if options.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Running in debug mode")
	}

	house := libplum.NewHouse()
	if options.Test {
		logrus.Debug("Running in test mode")
		house = libplum.NewTestHouse()
	}

	// if we have a state file, load it
	if file, err := os.Open(options.StateFile); err == nil {
		contents, err := ioutil.ReadAll(file)
		if err != nil {
			logrus.WithField("error", err).Error("failed to read state file")
		}
		house.LoadState(contents)
		// spew.Dump(house)
		logrus.WithField("house", house).Debug("loaded house config from state file")
		// and spin up a goroutine to save it in the background periodically
		go func() {
			tick := time.NewTicker(30 * time.Second).C
			for {
				select {
				case <-tick:
					contents, err = house.SaveState()
					if err != nil {
						logrus.WithField("error", err).Error("failed to get serialized house state")
					}
					err = ioutil.WriteFile(options.StateFile, contents, 0644)
					if err != nil {
						logrus.WithField("error", err).Error("failed to write state file")
					}
				}
			}
		}()
	}

	if options.Email != "" && options.Password != "" {
		house.SetCreds(&libplum.Account{
			Email:    options.Email,
			Password: options.Password,
		})
	}

	err := house.Initialize()
	if err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Second)
	house.Update()
	time.Sleep(5 * time.Second)
	// spew.Dump(house)
	nook, err := house.GetLoadByName("Nook")
	if err, ok := err.(libplum.ENotFound); ok {
		fmt.Println(err)
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}
	nook.SetTrigger(actions.OnMotionDetect(nook, 255))
	nook.SetTrigger(actions.OffAfterOn(nook, 20*time.Second))
	// go actions.OffAfterResetMotion(context.Background(), nook, 15*time.Second)
	for {
		fmt.Printf(".")
		time.Sleep(5 * time.Second)
	}
}
