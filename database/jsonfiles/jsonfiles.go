package jsonfiles

type Database struct {
}

type Bucket struct {
}

type Data struct {
}

func (db Database) Open(name string) error {
	return nil
}

func (db Database) Close() error {
	return nil
}

func (db Database) List() ([]Bucket, error) {
	return nil, nil
}

func (b Bucket) Name() string {
	return ""
}

func (b Bucket) List() ([]string, error) {
	return []string{}, nil
}

func (b Bucket) Save(ID string, value Data) error {
	return nil
}

func (b Bucket) Load(ID string, value *Data) error {
	return nil
}

func (b Bucket) String() string {
	return ""
}

func (d Data) String() string {
	return ""
}
