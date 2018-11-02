package services

type Service interface {
	Open() error
	Close() error
}
