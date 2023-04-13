package db

import "time"

type Model struct {
	ID           uint
	CreatedTime  time.Time
	ModifiedTime time.Time
}

type User struct {
	Model
	OpenId     string
	NickName   string
	LastSeen   time.Time
	LoginTimes uint
}

type Record struct {
	Model
	Uid    uint
	Type   uint
	Input  string
	Output string
}
