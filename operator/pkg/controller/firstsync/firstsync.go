package firstsync

type FirstSynchronizer interface {
	FirstSync() error
}
