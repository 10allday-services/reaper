package reapable

import (
	"fmt"
	"sync"

	"github.com/milescrabill/reaper/state"
)

type Terminable interface {
	Terminate() (bool, error)
}

type Stoppable interface {
	Stop() (bool, error)
	ForceStop() (bool, error)
}

type Whitelistable interface {
	Whitelist() (bool, error)
}

type Saveable interface {
	Save(state *state.State) (bool, error)
	Unsave() (bool, error)
	ReaperState() *state.State
	IncrementState() bool
}

//                ,____
//                |---.\
//        ___     |    `
//       / .-\  ./=)
//      |  | |_/\/|
//      ;  |-;| /_|
//     / \_| |/ \ |
//    /      \/\( |
//    |   /  |` ) |
//    /   \ _/    |
//   /--._/  \    |
//   `/|)    |    /
//     /     |   |
//   .'      |   |
//  /         \  |
// (_.-.__.__./  /
// credit: jgs, http://www.chris.com/ascii/index.php?art=creatures/grim%20reapers

type Reapable interface {
	Terminable
	Stoppable
	Whitelistable
	Saveable

	ReapableDescription() string
	ReapableDescriptionShort() string
	ReapableDescriptionTiny() string
}

type Reapables struct {
	sync.RWMutex
	storage map[string]map[string]Reapable
}

func NewReapables(regions []string) *Reapables {
	r := Reapables{}
	r.Lock()
	defer r.Unlock()

	// initialize Reapables map
	r.storage = make(map[string]map[string]Reapable)
	for _, region := range regions {
		r.storage[region] = make(map[string]Reapable)
	}
	return &r
}

func (rs *Reapables) Put(region, id string, r Reapable) {
	rs.Lock()
	defer rs.Unlock()
	rs.storage[region][id] = r
}

func (rs *Reapables) Get(region, id string) (Reapable, error) {
	rs.RLock()
	defer rs.Unlock()
	r, ok := rs.storage[region][id]
	if ok {
		return r, nil
	}
	return r, fmt.Errorf("Could not find %s", r.ReapableDescriptionTiny())
}

func (rs *Reapables) Delete(region, id string) {
	rs.RLock()
	defer rs.Unlock()
	delete(rs.storage[region], id)
}

type ReapableContainer struct {
	Reapable
	Region string
	ID     string
}

func (rs *Reapables) Iter() <-chan ReapableContainer {
	ch := make(chan ReapableContainer)
	go func(c chan ReapableContainer) {
		rs.Lock()
		defer rs.Unlock()
		for region, regionMap := range rs.storage {
			for id, r := range regionMap {
				c <- ReapableContainer{r, region, id}
			}
		}
	}(ch)
	return ch
}

// used to identify unowned resources
type UnownedError struct {
	ErrorText string
}

func (u UnownedError) Error() string {
	return u.ErrorText
}
