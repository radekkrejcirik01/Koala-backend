package notifications

type NotificationLike struct {
	Id             uint `gorm:"primary_key;auto_increment;not_null"`
	Sender         string
	NotificationId int
}

func (NotificationLike) TableName() string {
	return "notifications_likes"
}
