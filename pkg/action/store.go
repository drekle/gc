package action

/*
The assignment is to write a small library class that can perform the following operations:

1. Add Action
addAction (string) returning error
This function accepts a json serialized string of the form below and maintains an average time
for each action. 3 sample inputs:
1) {"action":"jump", "time":100}
2) {"action":"run", "time":75}
3) {"action":"jump", "time":200}
Assume that an end user will be making concurrent calls into this function.
2. Statistics
getStats () returning string
Write a second function that accepts no input and returns a serialized json array of the average
time for each action that has been provided to the addAction function. Output after the 3
sample calls above would be:
[
{"action":"jump", "avg":150},
{"action":"run", "avg":75}
]
Assume that an end user will be making concurrent calls into all functions.
*/
import (
	"encoding/json"
	"log"
	"sync"
)

var (
	// A configurable size of the buffered work queue
	QueueSize = 128
)

type Requirements interface {
	AddAction(json string) error
	GetStats() string
}

type actionStore struct {
	mapMutex *sync.Mutex
	actions  map[string]*actionStats
	work     chan *action
}

type action struct {
	Name string `json:"action"`
	// Time is not specified as Duration, custom marshalling could convert
	// Time has no known specified limit, so 64 bit
	// Time was not provided as a certain unit
	Time int64 `json:"time"`
}

type output struct {
	Name    string `json:"action"`
	Average int64  `json:"avg"`
}

type actionStats struct {
	avg int64
}

// This is excessive but enforces the requirements
var _ Requirements = &actionStore{}

func NewActionStore(stop chan struct{}) Requirements {
	processor := make(chan *action, QueueSize)
	store := &actionStore{
		actions: make(map[string]*actionStats),
		work:    processor,
	}

	go func() {
		for {
			select {
			case <-stop:
				break
			case a := <-processor:
				// The problem statment said nothing of storing the original values
				stats := store.getOrCreateStats(a)
				stats.avg = int64((stats.avg + a.Time) / 2)
			default:
				// continue
			}
		}
	}()

	return store
}

func (s *actionStore) AddAction(jsonData string) (err error) {
	var action action
	err = json.Unmarshal([]byte(jsonData), &action)
	if err != nil {
		return
	}
	s.work <- &action
	return
}

func (s *actionStore) GetStats() string {
	// the length of actions may change in a separate thread
	var ret []*output
	for name, stat := range s.actions {
		ret = append(ret, &output{
			Name:    name,
			Average: stat.avg,
		})
	}
	data, err := json.MarshalIndent(ret, "", "\t")
	if err != nil {
		// according to specified function prototype no way to handle error
		log.Printf("failed to marshal output: %v\n", err)
		return ""
	}

	return string(data)
}

func (s *actionStore) getOrCreateStats(a *action) (stats *actionStats) {
	var ok bool
	if stats, ok = s.actions[a.Name]; ok {
		return
	}
	stats = &actionStats{
		avg: a.Time,
	}
	s.actions[a.Name] = stats
	return stats
}
