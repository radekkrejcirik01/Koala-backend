package invites

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

type Invite struct {
	Id       uint   `gorm:"primary_key;auto_increment;not_null"`
	Sender   string `gorm:"size:256;index:idx_invites_sender_receiver"`
	Receiver string `gorm:"size:256;index:idx_invites_sender_receiver;index"`
	Accepted int
}

func (Invite) TableName() string {
	return "invites"
}

type InviteData struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	ProfilePhoto string `json:"profilePhoto,omitempty"`
}

func SendInvite(db *gorm.DB, t *Invite) (string, error) {
	if t.Sender == t.Receiver {
		return "Why are you inviting yourself? 😀", nil
	}

	var user users.UserData
	if err := db.
		Table("users").
		Select("id, username").
		Where("username = ?", t.Receiver).
		Find(&user).
		Error; err != nil {
		return "", err
	}

	if len(user.Username) == 0 {
		return "There is no user with this username", nil
	}

	// Rewrite case-insensitive username with db username
	t.Receiver = user.Username

	var invite Invite
	if err := db.
		Table("invites").
		Where("(sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)",
			t.Sender, t.Receiver, t.Receiver, t.Sender).
		Limit(1).
		Find(&invite).
		Error; err != nil {
		return "", err
	}

	if invite.Sender == t.Sender {
		return "This user was already invited", nil
	}
	if invite.Sender == t.Receiver {
		return "This user already invited you", nil
	}

	newInvite := Invite{
		Sender:   t.Sender,
		Receiver: t.Receiver,
	}

	if err := db.Table("invites").Create(&newInvite).Error; err != nil {
		return "", err
	}

	sender, err := users.GetUser(db, t.Sender)
	if err != nil {
		return "", err
	}

	tokens, err := service.GetTokensByUsername(db, t.Receiver)
	if err != nil {
		return "", err
	}

	fcmNotification := service.FcmNotification{
		Title:   sender.Name,
		Body:    "Sent you a friend invite",
		Sound:   "default",
		Devices: tokens,
	}

	return "Invite sent ✅", service.SendNotification(&fcmNotification)
}

// AcceptInvite updates accepted column in invites table
func AcceptInvite(db *gorm.DB, t *Invite) (string, error) {
	if err := db.
		Table("invites").
		Where("receiver = ? AND sender = ?", t.Sender, t.Receiver).
		Update("accepted", 1).
		Error; err != nil {
		return "", err
	}

	sender, err := users.GetUser(db, t.Sender)
	if err != nil {
		return "", err
	}

	tokens, err := service.GetTokensByUsername(db, t.Receiver)
	if err != nil {
		return "", err
	}

	fcmNotification := service.FcmNotification{
		Title:   sender.Name,
		Body:    "Accepted your friend invite",
		Sound:   "default",
		Devices: tokens,
	}

	return "", service.SendNotification(&fcmNotification)
}

// GetFriends gets friend from invites
func GetFriends(db *gorm.DB, username string) ([]users.UserData, error) {
	var invites []Invite

	if err := db.
		Table("invites").
		Select("sender, receiver").
		Where("(sender = ? OR receiver = ?) AND accepted = 1", username, username).
		Find(&invites).
		Error; err != nil {
		return nil, err
	}

	usernames := GetUsernamesFromInvites(invites, username)

	if len(usernames) == 0 {
		return nil, nil
	}

	var usersArray []users.UserData
	if err := db.
		Table("users").
		Select("id, username, name, profile_photo").
		Where("username IN ?", usernames).
		Find(&usersArray).Error; err != nil {
		return nil, err
	}

	usersMap := make(map[string]users.UserData, len(usersArray))
	for _, user := range usersArray {
		usersMap[user.Username] = user
	}

	var usersData []users.UserData
	for _, username := range usernames {
		usersData = append(usersData, usersMap[username])
	}

	return usersData, nil
}

// GetInvites gets friend invites from invites table
func GetInvites(db *gorm.DB, username string) ([]InviteData, error) {
	var invites []Invite
	var usernames []string
	var data []InviteData

	if err := db.
		Table("invites").
		Select("sender").
		Where("receiver = ? AND accepted = 0", username).
		Find(&invites).
		Error; err != nil {
		return nil, err
	}

	if len(invites) == 0 {
		return nil, nil
	}

	for _, invite := range invites {
		usernames = append(usernames, invite.Sender)
	}

	if err := db.
		Table("users").
		Select("id, username, profile_photo").
		Where("username IN ?", usernames).
		Find(&data).
		Error; err != nil {
		return nil, err
	}

	return data, nil
}

// RemoveFriend remove invite from invites table
func RemoveFriend(db *gorm.DB, id string, username string) error {
	var user string

	if err := db.
		Table("users").
		Select("username").
		Where("id = ?", id).
		Find(&user).
		Error; err != nil {
		return err
	}

	return db.
		Table("invites").
		Where("(sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)",
			user, username, username, user).
		Delete(&Invite{}).
		Error
}

// RemoveInvite remove invite form invites table, id is the sender's user id
func RemoveInvite(db *gorm.DB, id string, username string) error {
	var sender string

	if err := db.
		Table("users").
		Select("username").
		Where("id = ?", id).
		Find(&sender).
		Error; err != nil {
		return err
	}

	return db.
		Table("invites").
		Where("sender = ? AND receiver = ? AND accepted = 0", sender, username).
		Delete(&Invite{}).
		Error
}

// Helper function to get usernames from invites
func GetUsernamesFromInvites(invites []Invite, username string) []string {
	var usernames []string

	for _, invite := range invites {
		if invite.Sender == username {
			usernames = append(usernames, invite.Receiver)
		} else {
			usernames = append(usernames, invite.Sender)
		}
	}

	return usernames
}
