package main

type splitEvent struct {
	Kind, Target, Temporary string
	Success                 bool
}

type splitObserver func(splitEvent)

func emitSplit(observer splitObserver, kind, target, temporary string, success bool) {
	if observer != nil {
		observer(splitEvent{Kind: kind, Target: target, Temporary: temporary, Success: success})
	}
}
