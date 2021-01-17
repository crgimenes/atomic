package database

type Database interface {
	List() ([]Bucket, error)
}

type Bucket interface {
	Name() string
	List() ([]string, error)
	Save(ID string, value Data) error
	Load(ID string, value Data) error
	String() string
}

type Data interface {
	String() string
}
