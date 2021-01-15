package database

type Database interface {
	Open(name string) error
	Close() error
}
