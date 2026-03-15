package domain

import "time"

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Done       Status = "done"
)

type Task struct {
	Id          int
	Taskmsg     string
	Status      Status
	Createdat   time.Time
	Completedat time.Time
}
