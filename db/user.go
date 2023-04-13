package db

import (
	"log"
	"time"
)

type UserMapper struct {
}

func NewUserMapper() UserMapper {
	return UserMapper{}
}

func (mapper *UserMapper) Insert(openId string) (uint, error) {
	user := User{}
	if result := dbInstance.Where("open_id = ?", openId).First(&user); result.RowsAffected != 0 {
		log.Printf("user %s login, last seen at %s, total login times is now %d", openId, user.LastSeen, user.LoginTimes+1)
		// record info
		user.LastSeen = time.Now()
		user.LoginTimes = user.LoginTimes + 1
		dbInstance.Save(&user)
	} else {
		user = User{
			OpenId:     openId,
			LoginTimes: 1,
			LastSeen:   time.Now(),
		}
		user.CreatedTime = time.Now()
		user.ModifiedTime = time.Now()
		if result := dbInstance.Create(&user); result.RowsAffected == 0 {
			log.Println("create user failed, try next time, the error is: ", result.Error)
			return 0, result.Error
		}
	}
	return user.ID, nil
}
