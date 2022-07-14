package test_data

type TestDTO struct {
	Id         int64  `db:"id"`
	Value      string `db:"value"`
	unexported string
}
