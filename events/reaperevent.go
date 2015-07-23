package events

import (
	"fmt"

	log "github.com/mozilla-services/reaper/reaperlog"
)

type ReaperEventConfig struct {
	EventReporterConfig

	Mode string
}

type ReaperEvent struct {
	Config *ReaperEventConfig
}

func (e *ReaperEvent) SetDryRun(b bool) {
	e.Config.DryRun = b
}

func (*ReaperEvent) Cleanup() error { return nil }

func NewReaperEvent(c *ReaperEventConfig) *ReaperEvent {
	c.Name = "ReaperEvent"
	return &ReaperEvent{c}
}

func (*ReaperEvent) NewEvent(title string, text string, fields map[string]string, tags []string) error {
	return nil
}
func (*ReaperEvent) NewStatistic(name string, value float64, tags []string) error {
	return nil
}
func (*ReaperEvent) NewCountStatistic(name string, tags []string) error {
	return nil
}
func (e *ReaperEvent) NewReapableEvent(r Reapable) error {
	if e.Config.ShouldTriggerFor(r) {
		if log.Extras() {
			log.Error("Triggering ReaperEvent for %s", r.ReaperState().String())
		}
		var err error
		switch e.Config.Mode {
		case "Stop":
			_, err = r.Stop()
		case "Terminate":
			_, err = r.Terminate()
		default:
			log.Error(fmt.Sprintf("Invalid %s Mode %s", e.Config.Name, e.Config.Mode))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ReaperEvent) NewBatchReapableEvent(rs []Reapable) error {
	for _, r := range rs {
		err := e.NewReapableEvent(r)
		if err != nil {
			return err
		}
	}
	return nil
}
