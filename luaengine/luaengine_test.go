package luaengine

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/crgimenes/atomic/client"
)

type mockRead struct {
}

func (m *mockRead) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock error Read")
}

func TestRun(t *testing.T) {
	l := New()
	s := strings.NewReader(`print "1"`)
	c := &client.Instance{}
	err := l.InitState(s, c)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunError(t *testing.T) {
	l := New()
	s := strings.NewReader(`print 1`)
	c := &client.Instance{}
	err := l.InitState(s, c)
	if err == nil {
		t.Fatal("expecting error but got nil")
	}
}

func TestRunErrorReader(t *testing.T) {
	l := New()
	r := &mockRead{}
	c := &client.Instance{}
	err := l.InitState(r, c)
	if err == nil {
		t.Fatal("expecting error but got nil")
	}
}

func TestFileExists(t *testing.T) {
	file, err := ioutil.TempFile("", "ehsim")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if !fileExists(file.Name()) {
		t.Fatalf("expected tmp file to exists")
	}
}