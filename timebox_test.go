package timebox

import (
	"testing"

	"reflect"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/clock/fakeclock"
	"github.com/davars/timebox/internal/testproto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestSealAndReopen(t *testing.T) {
	Clock = fakeclock.NewFakeClock(time.Time{})
	defer func() {
		Clock = clock.NewClock()
	}()

	tests := map[string]struct {
		message proto.Message
		sealed  string
	}{
		"empty": {
			message: &emptypb.Empty{},
			sealed:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA_lhDK9z2adhnXlIn15w9asw1s39sPQ20FSEII0M",
		},
		"OAuthState empty": {
			message: &testproto.OAuthState{},
			sealed:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA_lhDK9z2adhnXlIn15w9asw1s39sPQ20FSEII0M",
		},
		"OAuthState gibberish": {
			message: &testproto.OAuthState{RedirectUrl: "asdf"},
			sealed:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAi3cP9pd4IJd3NLSVZu4Hr8w1s39sPQ20FSEII0Pmkl1pZM6z3w",
		},
		"OAuthState long": {
			message: &testproto.OAuthState{RedirectUrl: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAi3cP9pd4IJd3NLSVZu4Hr8w1s39sPQ20FSEII0Pmkl1pZM6z3w"},
			sealed:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFqCunyMTLmKbhvKLYgy5M8w1s39sPQ20FSEII0PmwF0_RPyW-Giq8QSz5tK2AWY8RALWQ5b-qypFWPb2kWu2AfUbu7HaqaT5hno8SfcrrvovRI6Ger1082-DByY9ax4BfkGCFdMmMLcTUJ7FhxgBVQ9zZQ",
		},
		"Session empty": {
			message: &testproto.Session{Authorized: false},
			sealed:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA_lhDK9z2adhnXlIn15w9asw1s39sPQ20FSEII0M",
		},
		"Session full": {
			message: &testproto.Session{Authorized: true, User: "somebody"},
			sealed:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANo6lli6oXCN6eIfWuFbtY8w1s39sPQ20FSEII0PmmF1ldtK63EuE1Dzqpg",
		},
	}

	b := Boxer{
		noncer: func() [24]byte {
			return [24]byte{}
		},
		secret: [32]byte{},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sealed, err := b.Seal(test.message, 0)
			assert.NoError(t, err)
			assert.Equal(t, test.sealed, sealed)

			v := reflect.New(reflect.TypeOf(test.message).Elem()).Interface().(proto.Message)
			assert.True(t, b.Open(sealed, v), "message opens")

			if !proto.Equal(v, test.message) {
				t.Errorf("got %+v, want %+v", v, test.message)
			}
		})
	}
}

func TestSealFailures(t *testing.T) {
	tests := map[string]struct {
		message proto.Message
		maxAge  int
		now     time.Time
	}{
		"nil": {
			message: nil,
		},
		"future": {
			message: &emptypb.Empty{},
			now:     time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	b := Boxer{
		noncer: func() [24]byte {
			return [24]byte{}
		},
		secret: [32]byte{},
	}

	defer func() {
		Clock = clock.NewClock()
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			Clock = fakeclock.NewFakeClock(test.now)
			sealed, err := b.Seal(test.message, time.Duration(test.maxAge)*time.Second)
			assert.Error(t, err)
			t.Log(err)
			assert.Equal(t, "", sealed)
		})
	}
}

func TestOpenFailures(t *testing.T) {
	tests := map[string]struct {
		message proto.Message
		sealed  string
		now     time.Time
	}{
		"empty": {
			sealed: "",
		},
		"short": {
			sealed: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		"not decryptable": {
			sealed: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAasfsfsaf",
		},
		"not a proto": {
			sealed: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAS610yvaV-aHUi0pmr6i7hqhRz9-fpb5ehKqY",
		},
		"invalid NotAfter": {
			sealed: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAyLhzIoaRpaXteTb-9-GGh8w4q39qbhIv",
		},
		"expired": {
			sealed: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA_lhDK9z2adhnXlIn15w9asw1s39sPQ20FSEII0M",
			now:    time.Time{}.Add(1),
		},
	}

	b := Boxer{
		noncer: func() [24]byte {
			return [24]byte{}
		},
		secret: [32]byte{},
	}

	defer func() {
		Clock = clock.NewClock()
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			Clock = fakeclock.NewFakeClock(test.now)

			v := &emptypb.Empty{}
			assert.False(t, b.Open(test.sealed, v))
		})
	}
}
