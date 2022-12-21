package model

import "time"

var (
	PAY_OUT int8 = 0
	PAY_IN  int8 = 1
)

type Bill struct {
	ID        int64     `gorm:"primaryKey,column:id" json:"id"`
	Gid       string    `gorm:"index:bgid_inx" json:"gid"`
	Created   time.Time `gorm:"index:bc_inx" json:"created"`
	Amount    float64   `gorm:"type:decimal(11,5);column:amount" json:"amount"`
	Order     string    `gorm:"column:order" json:"order"`
	Operator  string    `gorm:"column:operator" json:"operator"`
	Direction int8      `gorm:"column:direction" json:"direction"` //0:out 1:in
}

func (Bill) TableName() string {
	return "bills"
}
