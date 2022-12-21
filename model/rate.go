package model

import (
	"time"
)

type Rate struct {
	ID       int64     `gorm:"primaryKey,column:id" json:"id"`
	Gid      string    `gorm:"index:rgid_inx" json:"gid"`
	Created  time.Time `gorm:"index:rc_inx" json:"created"`
	InRate   float32   `gorm:"type:decimal(5,4);column:in_rate" json:"in_rate"`
	OutRate  float32   `gorm:"type:decimal(5,4);column:out_rate" json:"out_rate"`
	Operator string    `gorm:"column:operator" json:"operator"`
}

func (Rate) TableName() string {
	return "rates"
}
