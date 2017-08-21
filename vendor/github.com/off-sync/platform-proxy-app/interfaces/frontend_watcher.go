package interfaces

// FrontendEvent contains the information related to an event that happened
// to a frontend.
type FrontendEvent struct {
	Name string
}

// FrontendWatcher defines an interface for a frontend watcher against which
// it is possible to subscribe for frontend changes.
type FrontendWatcher interface {
	FrontendRepository

	// Subscribe returns a channel through which frontend events will
	// be distributed.
	Subscribe() <-chan FrontendEvent
}
