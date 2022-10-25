package engine

import (
	"sync"
)

//go:generate go run github.com/dmarkham/enumer -type=ProcessPriority -output enums.go

type ProgressCallbackFunc func(progress int, totalprogress int)

type ProcessorFunc func(ao *Objects)

type ProcessPriority int

const (
	BeforeMergeLow ProcessPriority = iota
	BeforeMerge
	BeforeMergeHigh
	BeforeMergeFinal
	AfterMergeLow
	AfterMerge
	AfterMergeHigh
	AfterMergeFinal
)

type ppfInfo struct {
	pf          ProcessorFunc
	description string
	priority    ProcessPriority
	loader      LoaderID
}

var registeredProcessors []ppfInfo

func (l LoaderID) AddProcessor(pf ProcessorFunc, description string, priority ProcessPriority) {
	registeredProcessors = append(registeredProcessors, ppfInfo{
		loader:      l,
		description: description,
		pf:          pf,
		priority:    priority,
	})
}

// LoaderID = wildcard
func Process(ao *Objects, cb ProgressCallbackFunc, l LoaderID, priority ProcessPriority) error {
	var priorityProcessors []ppfInfo
	for _, potentialProcessor := range registeredProcessors {
		if potentialProcessor.loader == l || l == -1 {
			if potentialProcessor.priority == priority {
				priorityProcessors = append(priorityProcessors, potentialProcessor)
			}
		}
	}

	aoLen := ao.Len()
	total := len(priorityProcessors) * aoLen

	if total == 0 {
		return nil
	}

	// We need to process this many objects
	cb(0, total)

	var wg sync.WaitGroup

	for _, processor := range priorityProcessors {
		wg.Add(1)
		go func(ppf ppfInfo) {
			ppf.pf(ao)
			cb(-aoLen, 0)
			wg.Done()
		}(processor)
	}
	wg.Wait()

	return nil
}
