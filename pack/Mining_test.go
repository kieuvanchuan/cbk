package pack

import (
	"reflect"
	"testing"
)

func TestMine(t *testing.T) {
	type args struct {
		tran  []Transaction
		phash [32]byte
	}
	tests := []struct {
		name string
		args args
		want Block
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Mine(tt.args.tran, tt.args.phash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPOW(t *testing.T) {
	type args struct {
		b    [32]byte
		diff int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"test_1",
			args{
				[32]byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				4,
			},
			true,
		},
		{
			"test_2",
			args{
				[32]byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				5,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPOW(tt.args.b, tt.args.diff); got != tt.want {
				t.Errorf("CheckPOW() = %v, want %v", got, tt.want)
			}
		})
	}
}
