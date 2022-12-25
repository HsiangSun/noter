package model

type Currency struct {
	ID       int64  `gorm:"primaryKey,column:id" json:"id"`
	Gid      string `gorm:"column:gid" json:"gid"`
	Currency string `gorm:"column:currency" json:"currency"`
}

func (Currency) TableName() string {
	return "currencies"
}
