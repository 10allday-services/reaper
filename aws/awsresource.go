package aws

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/mozilla-services/reaper/filters"
	"github.com/mozilla-services/reaper/reapable"
	log "github.com/mozilla-services/reaper/reaperlog"
	"github.com/mozilla-services/reaper/state"
)

// basic AWS resource, has properties that most/all resources have
type AWSResource struct {
	ID                 reapable.ID
	Name               string
	Region             reapable.Region
	Dependency         bool
	IsInCloudformation bool

	Tags map[string]string

	// reaper state
	reaperState *state.State

	// filters for MatchedFilters
	matchedFilterGroups map[string]filters.FilterGroup
}

func (a *AWSResource) Tagged(tag string) bool {
	_, ok := a.Tags[tag]
	return ok
}

// Tag returns the tag's value or an empty string if it does not exist
func (a *AWSResource) Tag(t string) string { return a.Tags[t] }

func (a *AWSResource) Owned() bool {
	// if the resource has an owner tag or a default owner is specified
	return a.Tagged("Owner") || config.DefaultOwner != ""
}

func (a *AWSResource) ReaperState() *state.State {
	return a.reaperState
}

func (a *AWSResource) SetReaperState(newState *state.State) {
	a.reaperState = newState
}

// Owner extracts useful information out of the Owner tag which should
// be parsable by mail.ParseAddress
func (a *AWSResource) Owner() *mail.Address {
	// properly formatted email
	if addr, err := mail.ParseAddress(a.Tag("Owner")); err == nil {
		return addr
	}

	// username -> default email host email address
	if addr, err := mail.ParseAddress(fmt.Sprintf("%s@%s", a.Tag("Owner"), config.DefaultEmailHost)); a.Tagged("Owner") && config.DefaultEmailHost != "" && err == nil {
		return addr
	}

	// default owner is specified
	if addr, err := mail.ParseAddress(
		fmt.Sprintf("%s@%s", config.DefaultOwner, config.DefaultEmailHost)); config.DefaultOwner != "" && config.DefaultEmailHost != "" && err == nil {
		return addr
	}
	log.Warning("No default owner or email host.")
	return nil
}

func (a *AWSResource) IncrementState() bool {
	var newState state.StateEnum
	until := time.Now()

	// did we update state?
	updated := false

	switch a.reaperState.State {
	default:
		// shouldn't ever be hit, but if it is
		// set state to the FirstState
		newState = state.FirstState
		until = until.Add(config.Notifications.FirstStateDuration.Duration)
	case state.FirstState:
		// go to SecondState at the end of FirstState
		newState = state.SecondState
		until = until.Add(config.Notifications.SecondStateDuration.Duration)
	case state.SecondState:
		// go to ThirdState at the end of SecondState
		newState = state.ThirdState
		until = until.Add(config.Notifications.ThirdStateDuration.Duration)
	case state.ThirdState:
		// go to FinalState at the end of ThirdState
		newState = state.FinalState
	case state.FinalState:
		// keep same state
		newState = state.FinalState
	case state.IgnoreState:
		// keep same state
		newState = state.IgnoreState
	}

	if newState != a.reaperState.State {
		updated = true
		log.Notice("Updating state for %s. New state: %s.", a.ReapableDescriptionTiny(), newState.String())
	}

	a.reaperState = state.NewStateWithUntilAndState(until, newState)

	return updated
}

func (a *AWSResource) AddFilterGroup(name string, fs filters.FilterGroup) {
	if a.matchedFilterGroups == nil {
		a.matchedFilterGroups = make(map[string]filters.FilterGroup)
	}
	a.matchedFilterGroups[name] = fs
}

func (a *AWSResource) MatchedFilters() string {
	return filters.FormatFilterGroupsText(a.matchedFilterGroups)
}

func (a *AWSResource) ReapableDescription() string {
	return fmt.Sprintf("%s matched %s", a.ReapableDescriptionShort(), a.MatchedFilters())
}

func (a *AWSResource) ReapableDescriptionShort() string {
	ownerString := ""
	if owner := a.Owner(); owner != nil {
		ownerString = fmt.Sprintf(" (owned by %s)", owner)
	}
	nameString := ""
	if name := a.Tag("Name"); name != "" {
		nameString = fmt.Sprintf(" \"%s\"", name)
	}
	return fmt.Sprintf("'%s'%s%s in %s with state: %s", a.ID, nameString, ownerString, a.Region, a.ReaperState().String())
}

func (a *AWSResource) ReapableDescriptionTiny() string {
	return fmt.Sprintf("'%s' in %s", a.ID, a.Region)
}

func (a *AWSResource) Whitelist() (bool, error) {
	return Whitelist(string(a.Region), string(a.ID))
}

// methods for reapable interface:
func (a *AWSResource) Save(s *state.State) (bool, error) {
	log.Notice("Saving %s", a.ReapableDescriptionTiny())
	return TagReaperState(string(a.Region), string(a.ID), s)
}

func (a *AWSResource) Unsave() (bool, error) {
	log.Notice("Unsaving %s", a.ReapableDescriptionTiny())
	return UntagReaperState(string(a.Region), string(a.ID))
}

func Whitelist(region, id string) (bool, error) {
	whitelist_tag := config.WhitelistTag

	api := ec2.New(&aws.Config{Region: region})
	req := &ec2.CreateTagsInput{
		Resources: []*string{aws.String(id)},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String(whitelist_tag),
				Value: aws.String("true"),
			},
		},
	}

	_, err := api.CreateTags(req)

	describereq := &ec2.DescribeTagsInput{
		DryRun: aws.Boolean(false),
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(id)},
			},
			&ec2.Filter{
				Name:   aws.String("key"),
				Values: []*string{aws.String(whitelist_tag)},
			},
		},
	}

	output, err := api.DescribeTags(describereq)

	if *output.Tags[0].Value == whitelist_tag {
		log.Info("Whitelist successful.")
		return true, err
	}

	return false, err
}

func UntagReaperState(region, id string) (bool, error) {
	api := ec2.New(&aws.Config{Region: region})
	delreq := &ec2.DeleteTagsInput{
		DryRun:    aws.Boolean(false),
		Resources: []*string{aws.String(id)},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key: aws.String(reaperTag),
			},
		},
	}
	_, err := api.DeleteTags(delreq)
	if err != nil {
		return false, err
	}
	return true, err
}

func TagReaperState(region, id string, newState *state.State) (bool, error) {
	api := ec2.New(&aws.Config{Region: region})
	createreq := &ec2.CreateTagsInput{
		DryRun:    aws.Boolean(false),
		Resources: []*string{aws.String(id)},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String(reaperTag),
				Value: aws.String(newState.String()),
			},
		},
	}

	_, err := api.CreateTags(createreq)
	if err != nil {
		return false, err
	}

	describereq := &ec2.DescribeTagsInput{
		DryRun: aws.Boolean(false),
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(id)},
			},
			&ec2.Filter{
				Name:   aws.String("key"),
				Values: []*string{aws.String(reaperTag)},
			},
		},
	}

	output, err := api.DescribeTags(describereq)

	if *output.Tags[0].Value == newState.String() {
		return true, err
	}

	return false, err
}
