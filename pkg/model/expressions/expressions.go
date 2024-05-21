package expressions

import (
	"time"

	i "github.com/radekkrejcirik01/Koala-backend/pkg/model/invites"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

type Expression struct {
	Id         uint `gorm:"primary_key;auto_increment;not_null"`
	UserId     int64
	Expression string
	Time       int64 `gorm:"autoCreateTime"`
}

func (Expression) TableName() string {
	return "expressions"
}

type ExpressionsData struct {
	Id         int64  `json:"id"`
	UserId     int64  `json:"userId"`
	Expression string `json:"expression"`
	Name       string `json:"name"`
	Time       int64  `json:"time"`
}

// PostExpression new expression to expressions table
func PostExpression(db *gorm.DB, t *Expression, username string) error {
	var invites []i.Invite

	err := db.Transaction(func(tx *gorm.DB) error {
		var expression []Expression

		if err := tx.
			Table("expressions").
			Where("user_id = ?", t.UserId).
			Find(&expression).
			Error; err != nil {
			return err
		}

		if len(expression) == 0 {
			return tx.
				Table("expressions").
				Create(&t).
				Error
		}
		return tx.
			Table("expressions").
			Where("user_id = ?", t.UserId).
			Updates(map[string]interface{}{"expression": t.Expression, "time": time.Now().Unix()}).
			Error
	})
	if err != nil {
		return err
	}

	if err := db.
		Table("invites").
		Where("(sender = ? OR receiver = ?) AND accepted = 1", username, username).
		Find(&invites).
		Error; err != nil {
		return err
	}

	usernames := i.GetUsernamesFromInvites(invites, username)

	var tokens []string
	tokens, err = service.GetTokensByUsernames(db, usernames)
	if err != nil {
		return err
	}

	var name string
	if err := db.
		Table("users").
		Select("name").
		Where("id = ?", t.UserId).
		Find(&name).
		Error; err != nil {
		return nil
	}

	fcmNotification := service.FcmNotification{
		Title:   name + " updated status",
		Body:    t.Expression,
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

// GetExpressions gets expressions from expressions table
func GetExpressions(db *gorm.DB, username string) ([]ExpressionsData, string, error) {
	var data []ExpressionsData
	var invites []i.Invite
	var userExpression string

	if err := db.
		Table("invites").
		Where("(sender = ? OR receiver = ?) AND accepted = 1", username, username).
		Find(&invites).
		Error; err != nil {
		return nil, "", err
	}

	usernames := i.GetUsernamesFromInvites(invites, username)
	usernames = append(usernames, username)

	var friends []users.UserData
	if err := db.
		Table("users").
		Select("id, name, username").
		Where("username IN ?", usernames).
		Find(&friends).
		Error; err != nil {
		return nil, "", err
	}

	usersIds := getUserIds(friends)

	var expressions []Expression
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Unix()
	if err := db.
		Table("expressions").
		Where("user_id IN ? AND (time > ? OR time IS NULL)", usersIds, twoDaysAgo).
		Find(&expressions).
		Error; err != nil {
		return nil, "", err
	}

	for _, v := range expressions {
		if getUsername(v.UserId, friends) == username {
			userExpression = v.Expression
		} else {
			data = append(data, ExpressionsData{
				Id:         int64(v.Id),
				UserId:     v.UserId,
				Name:       getUserName(v.UserId, friends),
				Expression: v.Expression,
				Time:       v.Time,
			})
		}
	}

	return data, userExpression, nil
}

// ClearStatus remove expression from expressions table
func ClearStatus(db *gorm.DB, username string) error {
	var userId int64

	if err := db.
		Table("users").
		Select("id").
		Where("username = ?", username).
		Find(&userId).
		Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("expressions").
			Where("user_id = ?", userId).
			Delete(&Expression{}).
			Error
	})
}

func getUserIds(users []users.UserData) []int64 {
	var ids []int64
	for _, v := range users {
		ids = append(ids, v.Id)
	}
	return ids
}

func getUserName(userId int64, users []users.UserData) string {
	for _, v := range users {
		if v.Id == userId {
			return v.Name
		}
	}
	return ""
}

func getUsername(userId int64, users []users.UserData) string {
	for _, v := range users {
		if v.Id == userId {
			return v.Username
		}
	}
	return ""
}
