package db

import (
	"time"
)

type Model struct {
	ID           uint      // unique auto increased id
	CreatedTime  time.Time // created time
	ModifiedTime time.Time // modified time
}

type User struct {
	Model
	OpenId     string    // wechat user unique id
	NickName   string    // nickname
	LastSeen   time.Time // last seen time
	LoginTimes uint      // login times
}

type Record struct {
	Model
	Uid    uint   // user id
	Type   string // PROMPT or VARIATION
	Input  string // prompt text or variation origin image path
	Output string // generated image path
}

type Task struct {
	Model
	Uid    uint   // user id
	Type   string // task type, STABLE_DIFFUSION
	RawReq string // request json info
	Status string // task status, PENDING, RUNNING, SUCCEED, FAILED
	ErrMsg string // error message
}
