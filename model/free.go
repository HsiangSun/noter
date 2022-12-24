package model

import "time"

type Free struct {
	ID       int64     `gorm:"primaryKey,column:id" json:"id"`
	Gid      string    `gorm:"index:fgid_inx" json:"gid"`
	Created  time.Time `gorm:"index:rf_inx" json:"created"`
	InFree   float32   `gorm:"type:decimal(5,4);column:in_free" json:"in_free"`
	OutFree  float32   `gorm:"type:decimal(5,4);column:out_free" json:"out_free"`
	Operator string    `gorm:"column:operator" json:"operator"`
}

func (Free) TableName() string {
	return "frees"
}
