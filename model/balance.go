package model

import "time"

type Balance struct {
	ID      int64     `gorm:"primaryKey,column:id" json:"id"`
	Gid     string    `gorm:"index:gid_inx" json:"gid"`
	Created time.Time `gorm:"index:ct_idx" json:"created"`
	PayIn   float64   `gorm:"type:decimal(11,5);column:pay_in" json:"pay_in"`
	PayOut  float64   `gorm:"type:decimal(11,5);column:pay_out" json:"pay_out"`
}
