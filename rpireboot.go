package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/ardnew/version"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	commandName     = "rpireboot"
	defaultPin      = 18
	defaultPull     = "up"
	defaultEdge     = "fall"
	defaultDebounce = 5 * time.Millisecond
)

func init() {
	version.ChangeLog = []version.Change{{
		Package: commandName,
		Version: "0.1.0",
		Date:    "2020 Apr 10",
		Description: []string{
			"initial revision",
		},
	}}
}

func printChangeLog() {
	version.PrintChangeLog()
}

func printVersion() {
	fmt.Printf("%s version %s\n", commandName, version.String())
}

func main() {

	var (
		argChangeLog bool
		argVersion   bool
		argPin       uint
		argPull      string
		argEdge      string
		argDebounce  time.Duration
	)

	flag.BoolVar(&argChangeLog, "changelog", false, "display change history")
	flag.BoolVar(&argVersion, "version", false, "display version information")
	flag.UintVar(&argPin, "p", defaultPin, "listen for interrupts on GPIO pin `n`")
	flag.StringVar(&argPull, "l", defaultPull, "configure pull on interrupt pin as `up|down|float`")
	flag.StringVar(&argEdge, "e", defaultEdge, "trigger interrupt on detected edge `rise|fall|both`")
	flag.DurationVar(&argDebounce, "d", defaultDebounce, "debounce reads with `duration`")
	flag.Parse()

	if argChangeLog {
		printChangeLog()
	} else if argVersion {
		printVersion()
	} else {

		if _, err := host.Init(); err != nil {
			log.Fatal(err)
		}

		pinName := fmt.Sprintf("GPIO%d", argPin)

		queue, err := NewInterruptQueue(pinName, argPull, argEdge)
		if err != nil {
			log.Fatalf("failed to initialize GPIO interrupt: %v", err)
		}
		go queue.Watch()

		queue.Listen(argDebounce)
	}
}

func reboot() {
	syscall.Sync()
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

type (
	Interrupt struct {
		level gpio.Level
		when  time.Time
	}
	InterruptQueue struct {
		pin gpio.PinIO
		ich chan *Interrupt
	}
)

func parsePull(s string) (gpio.Pull, bool) {

	if re, err := regexp.Compile("^" + strings.TrimSpace(strings.ToLower(s))); err == nil {
		if re.MatchString("pullup") || re.MatchString("up") {
			return gpio.PullUp, true
		} else if re.MatchString("pulldown") || re.MatchString("down") {
			return gpio.PullDown, true
		} else if re.MatchString("float") || re.MatchString("none") {
			return gpio.Float, true
		}
	}
	return gpio.PullNoChange, false
}

func parseEdge(s string) (gpio.Edge, bool) {

	if re, err := regexp.Compile("^" + strings.TrimSpace(strings.ToLower(s))); err == nil {
		if re.MatchString("rise") || re.MatchString("rising") {
			return gpio.RisingEdge, true
		} else if re.MatchString("fall") {
			return gpio.FallingEdge, true
		} else if re.MatchString("both") {
			return gpio.BothEdges, true
		}
	}
	return gpio.NoEdge, false
}

func NewInterruptQueue(pinName string, pullName string, edgeName string) (*InterruptQueue, error) {

	pin := gpioreg.ByName(pinName)
	if pin == nil {
		return nil, fmt.Errorf("invalid pin name: %q", pinName)
	}

	pull, ok := parsePull(pullName)
	if !ok {
		return nil, fmt.Errorf("invalid pull name: %q", pullName)
	}

	edge, ok := parseEdge(edgeName)
	if !ok {
		return nil, fmt.Errorf("invalid edge name: %q", edgeName)
	}

	if err := pin.In(pull, edge); err != nil {
		return nil, fmt.Errorf("could not configure %s (dir=IN, pull=%s, edge=%s): %v",
			pinName, pullName, edgeName, err)
	}

	return &InterruptQueue{
		pin: pin,
		ich: make(chan *Interrupt),
	}, nil
}

func (q *InterruptQueue) Watch() {

	for {
		q.pin.WaitForEdge(-1)
		q.ich <- &Interrupt{
			level: q.pin.Read(),
			when:  time.Now(),
		}
	}
	close(q.ich) // error, channel should never close
}

func (q *InterruptQueue) Listen(debounce time.Duration) {

	var last *Interrupt = nil
	for {
		recv := <-q.ich
		valid := last == nil
		//if !valid && recv.level != last.level {
		if !valid {
			valid = recv.when.Sub(last.when) > debounce
		}
		if valid {
			last = recv
			reboot()
		}
	}
}
