package server

import (
	"reflect"
	"testing"

	"github.com/crgimenes/atomic/luaengine"
)

func TestNew(t *testing.T) {
	l := luaengine.New()
	sshServer := New(l)
	type args struct {
		le LuaEngine
	}
	tests := []struct {
		name string
		args args
		want *SSHServer
	}{
		{
			name: "sucess",
			args: args{
				le: l,
			},
			want: sshServer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.le); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
