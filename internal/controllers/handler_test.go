package controllers

import (
	"reflect"
	"testing"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog"
)

func TestNewHandler(t *testing.T) {
	logger := zerolog.Logger{}
	c := clamav.ClamavClient{}
	type args struct {
		logger *zerolog.Logger
		clamav clamav.Clamaver
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil, nil},
			want: &Handler{nil, nil},
		},
		{
			name: "non nil args",
			args: args{&logger, &c},
			want: &Handler{&c, &logger},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.logger, tt.args.clamav); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
