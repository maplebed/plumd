package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	flag "github.com/jessevdk/go-flags"
	"github.com/maplebed/libplum"
	"github.com/maplebed/plumd/actions"
)

type Options struct {
	Email     string `short:"e" long:"email" descrption:"Email address to authenticate with the Plum Web API"`
	Password  string `short:"p" long:"password" descrption:"Password to authenticate with the Plum Web API"`
	StateFile string `short:"f" long:"file" description:"location to store state file" default:"/var/lib/plumd.state"`
}

func main() {
	var options Options
	flagParser := flag.NewParser(&options, flag.Default)
	flagParser.Parse()

	house := libplum.PlumHouse{}

	// if we have a state file, load it
	if file, err := os.Open(options.StateFile); err == nil {
		contents, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		house.LoadState(contents)
	}

	// update email and password from flags
	if options.Email != "" {
		house.Email = options.Email
	}
	if options.Password != "" {
		house.Password = options.Password
	}

	house.Initialize("aoeu")

	nook := house.GetLoadByName("Nook")
	if nook == nil {
		panic("nook not found")
	}
	nook.SetTrigger(actions.OnMotionDetect(nook, 255))
	nook.SetTrigger(actions.OffAfterOn(nook, 20*time.Second))
	// go actions.OffAfterResetMotion(context.Background(), nook, 15*time.Second)
	for {
		fmt.Printf(".")
		time.Sleep(5 * time.Second)
	}
}
