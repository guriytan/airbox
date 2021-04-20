package dto

import "airbox/global"

type QueryCondition struct {
	StorageID int64
	FatherID  *int64
	Type      *global.FileType
	Cursor    int64
	Limit     int
}

func (cond *QueryCondition) IsSetFatherID() bool {
	return cond.FatherID != nil
}

func (cond *QueryCondition) GetFatherID() int64 {
	if cond.IsSetFatherID() {
		return *cond.FatherID
	}
	return 0
}

func (cond *QueryCondition) IsSetType() bool {
	return cond.Type != nil
}

func (cond *QueryCondition) GetType() global.FileType {
	if cond.IsSetType() {
		return *cond.Type
	}
	return 0
}
