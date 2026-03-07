package logparser

import (
	"reflect"
	"testing"
	"time"
)

func Test_newEntry(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		buf     []byte
		want    dataEntry
		wantErr bool
	}{
		{
			"test 1",
			[]byte("PRG chuwi 1771747662 2026/02/22 11:07:42 588957 1 (systemd) S 0 0 1 1 0 1771158705 (/sbin/init splash)"),
			dataEntry{
				label:     labelPRG,
				computer:  "chuwi",
				timeStamp: time.Date(2026, 02, 22, 11, 07, 42, 0, time.Local),
				interval:  588957,
				points: [][]byte{
					[]byte("1"),
					[]byte("systemd"),
					[]byte("S"),
					[]byte("0"),
					[]byte("0"),
					[]byte("1"),
					[]byte("1"),
					[]byte("0"),
					[]byte("1771158705"),
					[]byte("/sbin/init splash"),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := newEntry(tt.buf)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("newEntry() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("newEntry() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}
