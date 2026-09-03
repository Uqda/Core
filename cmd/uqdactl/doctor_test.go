package main

import (
	"testing"

	"github.com/Uqda/Core/src/admin"
)

func TestDoctorExitCode(t *testing.T) {
	tests := []struct {
		status string
		want   int
	}{
		{admin.DoctorPass, 0},
		{admin.DoctorWarn, 2},
		{admin.DoctorFail, 1},
	}
	for _, tt := range tests {
		if got := doctorExitCode(admin.DoctorResponse{Status: tt.status}); got != tt.want {
			t.Errorf("status %q: got %d, want %d", tt.status, got, tt.want)
		}
	}
}
