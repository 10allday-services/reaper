package events

import (
	"fmt"
	"strconv"
	"text/template"

	"github.com/PagerDuty/godspeed"
)

// DataDogConfig is the configuration for a DataDog
type DataDogConfig struct {
	Enabled bool
	DryRun  bool
	Host    string
	Port    string
}

// implements EventReporter, sends events and statistics to DataDog
// uses godspeed, requires dd-agent running
type DataDog struct {
	Config        *DataDogConfig
	eventTemplate template.Template
	godspeed      *godspeed.Godspeed
}

func NewDataDog(c *DataDogConfig) *DataDog {
	return &DataDog{
		Config: c,
	}
}

// TODO: make this async?
// TODO: don't recreate godspeed
func (d *DataDog) Godspeed() (*godspeed.Godspeed, error) {
	if d.godspeed == nil {
		var g *godspeed.Godspeed
		var err error
		// if config options not set, use defaults
		if d.Config.Host == "" || d.Config.Port == "" {
			g, err = godspeed.NewDefault()
		} else {
			port, err := strconv.Atoi(d.Config.Port)
			if err != nil {
				return nil, err
			}
			g, err = godspeed.New(d.Config.Host, port, false)
		}
		if err != nil {
			return nil, err
		}
		d.godspeed = g
	}
	return d.godspeed, nil
}

// NewEvent reports an event to DataDog
func (d *DataDog) NewEvent(title string, text string, fields map[string]string, tags []string) error {
	g, err := d.Godspeed()
	if err != nil {
		return err
	}
	// TODO: fix?
	// defer g.Conn.Close()
	err = g.Event(title, text, fields, tags)
	if err != nil {
		return fmt.Errorf("Error reporting Godspeed event %s: %s", title, err)
	}
	return nil
}

// NewStatistic reports a gauge to DataDog
func (d *DataDog) NewStatistic(name string, value float64, tags []string) error {
	g, err := d.Godspeed()
	if err != nil {
		return err
	}
	// TODO: fix?
	// defer g.Conn.Close()
	err = g.Gauge(name, value, tags)
	if err != nil {
		return fmt.Errorf("Error reporting Godspeed statistic %s: %s", name, err)
	}

	return nil
}

// NewCountStatistic reports an Incr to DataDog
func (d *DataDog) NewCountStatistic(name string, tags []string) error {
	g, err := d.Godspeed()
	if err != nil {
		return err
	}
	// TODO: fix?
	// defer g.Conn.Close()
	err = g.Incr(name, tags)
	if err != nil {
		return fmt.Errorf("Error reporting Godspeed count statistic %s: %s", name, err)
	}
	return nil
}

func (d *DataDog) NewReapableEvent(r Reapable) error {
	return nil
}
