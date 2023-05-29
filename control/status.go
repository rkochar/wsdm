package main

type Status string

const (
	Completed          Status = "Completed"
	Pending                   = "Pending"
	Rollback_Pending          = "Rollback_Pending"
	Rollback_Completed        = "Rollback_Completed"
	Failure                   = "Failure"
)
