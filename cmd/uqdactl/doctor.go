package main

import "github.com/Uqda/Core/src/admin"

func doctorExitCode(res admin.DoctorResponse) int {
	switch res.Status {
	case admin.DoctorFail:
		return 1
	case admin.DoctorWarn:
		return 2
	default:
		return 0
	}
}
