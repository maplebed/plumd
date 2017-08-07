package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/maplebed/libplum"
	"github.com/maplebed/libplumraw"
)

// Define functions here that you want to act as automated bits to your plumd.
// Define triggers, express behaviors, etc. Then add them to the main plumd to
// get run.

// OffAfter turns off a light some amount of time after it has been turned on.
func OffAfter(ctx context.Context, load libplum.LogicalLoad, dur time.Duration) {
	fmt.Printf("starting timer, will shut off at %s\n", time.Now().Add(dur))
	// selecting on the context so the caller can cancel the auto-off
	select {
	case <-time.After(dur):
		fmt.Printf("calling set level at %s\n", time.Now())
		load.SetLevel(0)
	case <-ctx.Done():
		fmt.Printf("got cancelled\n")
		// cancelling this autooff without turning off the light
		return
	}
}

// OffAfterResetMotion turns a light off after a certain amount of time and
// resets the timer each time motion is detected by the switch
func OffAfterResetMotion(ctx context.Context, load libplum.LogicalLoad, dur time.Duration) {
	newctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resetDelay := func(ev libplumraw.Event) {
		switch ev.(type) {
		case libplumraw.LPEPIRSignal:
			fmt.Printf("it's a PIR signal, resetting timer\n")
			cancel()
			OffAfterResetMotion(ctx, load, dur)
		}
	}
	load.SetTrigger(&resetDelay)
	OffAfter(newctx, load, dur)
	load.ClearTrigger(&resetDelay)
}

// OffAfterOn is a trigger to autamatically turn the light off some duration
// after it was turned on. Motion resets the countdown timer
func OffAfterOn(load libplum.LogicalLoad, dur time.Duration) libplum.TriggerFn {
	oao := func(ev libplumraw.Event) {
		switch ev := ev.(type) {
		case libplumraw.LPEDimmerChange:
			if ev.Level > 10 {
				fmt.Printf("Light turned on! starting timer...\n")
				OffAfterResetMotion(context.Background(), load, dur)
			}
		}
	}
	return &oao
}
