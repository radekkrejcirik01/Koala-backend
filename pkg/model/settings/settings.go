package settings

import (
	"encoding/json"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const NicknameLimit = 32
const ThemeLimit = 32

// Setting is user's private setting of the chat with a friend, the friend never sees it
type Setting struct {
	UserId   int64   `gorm:"primaryKey;autoIncrement:false"`
	FriendId int64   `gorm:"primaryKey;autoIncrement:false;index"`
	Nickname *string `gorm:"size:32"`
	Theme    *string `gorm:"size:32"`
}

func (Setting) TableName() string {
	return "chat_settings"
}

type SettingsData struct {
	Nickname *string `json:"nickname,omitempty"`
	Theme    *string `json:"theme,omitempty"`
}

// Optional tells a missing field, which stays unchanged, from null, which removes the value
type Optional struct {
	Set   bool
	Value *string
}

func (o *Optional) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	return json.Unmarshal(data, &o.Value)
}

type UpdateRequest struct {
	Nickname Optional `json:"nickname"`
	Theme    Optional `json:"theme"`
}

// Get gets user's settings of the chat with a friend, empty when nothing is set
func Get(db *gorm.DB, userId, friendId int64) (SettingsData, error) {
	var data SettingsData
	err := db.
		Table("chat_settings").
		Select("nickname, theme").
		Where("user_id = ? AND friend_id = ?", userId, friendId).
		Limit(1).
		Find(&data).
		Error

	return data, err
}

// Save saves the columns of the setting, other columns keep their value,
// a setting without any value is deleted
func Save(db *gorm.DB, setting Setting, columns []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Table("chat_settings").
			Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns(columns)}).
			Create(&setting).
			Error; err != nil {
			return err
		}

		return tx.
			Table("chat_settings").
			Where("user_id = ? AND friend_id = ? AND nickname IS NULL AND theme IS NULL",
				setting.UserId, setting.FriendId).
			Delete(&Setting{}).
			Error
	})
}

// SetNicknames sets the nicknames the user gave to the friends
func SetNicknames(db *gorm.DB, username string, friends []users.UserData) error {
	if len(friends) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(friends))
	for _, friend := range friends {
		ids = append(ids, friend.Id)
	}

	var settings []Setting
	if err := db.
		Table("chat_settings").
		Select("chat_settings.friend_id, chat_settings.nickname").
		Joins("JOIN users ON users.id = chat_settings.user_id").
		Where("users.username = ? AND chat_settings.friend_id IN ? AND chat_settings.nickname IS NOT NULL",
			username, ids).
		Find(&settings).
		Error; err != nil {
		return err
	}

	nicknames := make(map[int64]string, len(settings))
	for _, setting := range settings {
		nicknames[setting.FriendId] = *setting.Nickname
	}

	for i := range friends {
		friends[i].Nickname = nicknames[friends[i].Id]
	}

	return nil
}

// GetSenderNicknames gets the nicknames the users gave to the sender, keyed by user id
func GetSenderNicknames(db *gorm.DB, senderId int64, userIds []int64) (map[int64]string, error) {
	var settings []Setting
	if err := db.
		Table("chat_settings").
		Select("user_id, nickname").
		Where("friend_id = ? AND user_id IN ? AND nickname IS NOT NULL", senderId, userIds).
		Find(&settings).
		Error; err != nil {
		return nil, err
	}

	nicknames := make(map[int64]string, len(settings))
	for _, setting := range settings {
		nicknames[setting.UserId] = *setting.Nickname
	}

	return nicknames, nil
}

// DeleteUserSettings deletes user's settings and settings other users have of the chat with the user
func DeleteUserSettings(tx *gorm.DB, userId int64) error {
	return tx.
		Table("chat_settings").
		Where("user_id = ? OR friend_id = ?", userId, userId).
		Delete(&Setting{}).
		Error
}
