package client

import (
	"reflect"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestNewInstance(t *testing.T) {
	type args struct {
		conn ssh.Channel
	}
	tests := []struct {
		name string
		args args
		want *Instance
	}{
		{
			name: "success",
			args: args{
				conn: nil,
			},
			want: &Instance{
				Conn: nil,
				H:    25,
				W:    80,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInstance(tt.args.conn, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}
