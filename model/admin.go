package model

type Admin struct {
	ID  int64  `gorm:"primaryKey,column:id" json:"id"`
	Uid string `gorm:"column:uid" json:"uid"`
	Gid string `gorm:"column:gid" json:"gid"`
}

func (Admin) TableName() string {
	return "admins"
}
