package utils

import (
	"reflect"
	"testing"
)

func TestParseCPUList(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:    "Single CPU",
			input:   "0",
			want:    []int{0},
			wantErr: false,
		},
		{
			name:    "Multiple CPUs",
			input:   "0,1,2",
			want:    []int{0, 1, 2},
			wantErr: false,
		},
		{
			name:    "CPU Range",
			input:   "0-3",
			want:    []int{0, 1, 2, 3},
			wantErr: false,
		},
		{
			name:    "Mixed Format",
			input:   "0-2,4,6-8",
			want:    []int{0, 1, 2, 4, 6, 7, 8},
			wantErr: false,
		},
		{
			name:    "With Spaces",
			input:   " 0, 1 , 2-4 ",
			want:    []int{0, 1, 2, 3, 4},
			wantErr: false,
		},
		{
			name:    "Empty Input",
			input:   "",
			want:    []int{},
			wantErr: false,
		},
		{
			name:    "Invalid Range",
			input:   "3-1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid Number",
			input:   "a",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid Range Format",
			input:   "1-2-3",
			want:    nil,
			wantErr: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCPUList(tt.input)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCPUList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if result matches expectation
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCPUList() = %v, want %v", got, tt.want)
			}
		})
	}
}
