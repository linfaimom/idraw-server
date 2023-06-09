package db

import (
	"log"
	"time"
)

type RecordMapper struct {
}

func NewRecordMapper() RecordMapper {
	return RecordMapper{}
}

func (mapper *RecordMapper) Insert(openId string, calledType string, input string, output string) (uint, error) {
	user := User{}
	if result := dbInstance.Where("open_id = ?", openId).First(&user); result.RowsAffected == 0 {
		log.Printf("failed to find the user, the error is %s, skip recording", result.Error)
		return 0, result.Error
	}
	record := Record{
		Uid:    user.ID,
		Type:   calledType,
		Input:  input,
		Output: output,
	}
	record.CreatedTime = time.Now()
	record.ModifiedTime = time.Now()
	if result := dbInstance.Create(&record); result.RowsAffected == 0 {
		log.Println("create record failed, the error is: ", result.Error)
		return 0, result.Error
	}
	return record.ID, nil
}

func (mapper *RecordMapper) FetchCountByUser(openId string) (int, error) {
	user := User{}
	if result := dbInstance.Where("open_id = ?", openId).First(&user); result.RowsAffected == 0 {
		log.Printf("failed to find the user, the error is %s", result.Error)
		return 0, result.Error
	}
	records := []Record{}
	result := dbInstance.Where("uid = ?", user.ID).Find(&records)
	return len(records), result.Error
}

func (mapper *RecordMapper) FetchByUserAndType(openId string, calledType string) ([]Record, error) {
	user := User{}
	if result := dbInstance.Where("open_id = ?", openId).First(&user); result.RowsAffected == 0 {
		log.Printf("failed to find the user, the error is %s", result.Error)
		return []Record{}, result.Error
	}
	records := []Record{}
	result := dbInstance.Where("uid = ? and type = ?", user.ID, calledType).Order("modified_time desc").Find(&records)
	return records, result.Error
}
