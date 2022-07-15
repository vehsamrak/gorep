package test_data

import "time"

type TestDTO struct {
	Id    int64     `db:"id"`
	Value string    `db:"value"`
	Time  time.Time `db:"time"`
}
