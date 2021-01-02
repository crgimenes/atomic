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
	closed bool
	data   []byte
}

func (m *mockConn) Read(data []byte) (int, error) {
	return 0, nil
}

func (m *mockConn) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) CloseWrite() error {
	m.closed = true
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
	c := client.NewInstance(nil, l)
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

func TestQuit(t *testing.T) {
	l := New()
	s := strings.NewReader(`quit()`)
	m := &mockConn{}
	c := &client.Instance{
		Conn: m,
	}
	err := l.InitState(s, c)
	if err != nil {
		t.Fatal("running quit() function", err)
	}
	if !m.closed {
		t.Fatal("quit() function did not close the connection")
	}
}

func TestWrite(t *testing.T) {
	l := New()
	s := strings.NewReader(`write("test")`)
	m := &mockConn{}
	c := &client.Instance{
		Conn: m,
	}
	err := l.InitState(s, c)
	if err != nil {
		t.Fatal("running write() function", err)
	}
	if string(m.data) != "test" {
		t.Fatal("error on write() function")
	}
}

func TestCls(t *testing.T) {
	l := New()
	s := strings.NewReader(`cls()`)
	m := &mockConn{}
	c := &client.Instance{
		Conn: m,
	}
	err := l.InitState(s, c)
	if err != nil {
		t.Fatal("running cls() function", err)
	}
	if string(m.data) != "\u001b[2J" {
		t.Fatal("error on cls() function")
	}
}
