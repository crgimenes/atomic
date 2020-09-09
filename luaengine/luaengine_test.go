package luaengine

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/crgimenes/atomic/client"
	lua "github.com/yuin/gopher-lua"
)

type mockRead struct {
}

func (m *mockRead) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock error Read")
}

type mockConn struct {
}

func (m *mockConn) Read(data []byte) (int, error) {
	return 0, nil
}

func (m *mockConn) Write(data []byte) (int, error) {
	return 0, nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) CloseWrite() error {
	return nil
}

func (m *mockConn) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return false, nil
}

func (m *mockConn) Stderr() io.ReadWriter {
	return nil
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
	file, err := ioutil.TempFile("", "file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if !fileExists(file.Name()) {
		t.Fatalf("expected tmp file to exists")
	}
	if fileExists("file-not-exists") {
		t.Fatalf("expected file to exists")
	}

}

func TestRunPwd(t *testing.T) {
	l := New()
	c := client.NewInstance(nil)
	s := strings.NewReader(`pwdStr = pwd()`)
	err := l.InitState(s, c)
	if err != nil {
		t.Fatal(err)
	}
	L := l.GetState()
	resp := L.GetGlobal("pwdStr").(lua.LString).String()
	if resp != pwd() {
		t.Fatalf("expect %q but got %q", resp, pwd())
	}
}
