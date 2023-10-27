package invites

import (
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

type Invite struct {
	Id       uint `gorm:"primary_key;auto_increment;not_null"`
	Sender   string
	Receiver string
	Accepted int
}

func (Invite) TableName() string {
	return "invites"
}

func SendInvite(db *gorm.DB, t *Invite) (string, error) {
	if t.Sender == t.Receiver {
		return "Why are you inviting yourself? 😀", nil
	}

	var friendsNumber int64
	if err := db.
		Table("invites").
		Where("(sender = ? OR receiver = ?) AND accepted = 1", t.Receiver, t.Receiver).
		Count(&friendsNumber).
		Error; err != nil {
		return "", err
	}

	if friendsNumber >= 3 {
		return "This user has already added 3 people", nil
	}

	var user users.UserData
	if err := db.
		Table("users").
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

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("invites").Create(&newInvite).Error
	})
	if err != nil {
		return "", err
	}

	tokens, err := service.GetTokensByUsername(db, t.Receiver)
	if err != nil {
		return "", err
	}

	fcmNotification := service.FcmNotification{
		Title:   t.Sender,
		Body:    t.Sender + " is sending a friend invite",
		Sound:   "default",
		Devices: tokens,
	}

	return "Invite sent ✅", service.SendNotification(&fcmNotification)
}

// AcceptInvite updates accepted column in invites table
func AcceptInvite(db *gorm.DB, t *Invite) (string, error) {
	var friendsNumber int64
	if err := db.
		Table("invites").
		Where("(sender = ? OR receiver = ?) AND accepted = 1", t.Sender, t.Sender).
		Count(&friendsNumber).
		Error; err != nil {
		return "", err
	}

	if friendsNumber >= 3 {
		return "You have already added 3 people", nil
	}

	var acceptedNumber int64
	if err := db.
		Table("invites").
		Where("(sender = ? OR receiver = ?) AND accepted = 1", t.Receiver, t.Receiver).
		Count(&acceptedNumber).
		Error; err != nil {
		return "", err
	}

	if acceptedNumber >= 3 {
		return "This user has already added 3 people", nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("invites").
			Where("receiver = ? AND sender = ?", t.Sender, t.Receiver).
			Update("accepted", 1).
			Error
	})

	tokens, err := service.GetTokensByUsername(db, t.Receiver)
	if err != nil {
		return "", err
	}

	fcmNotification := service.FcmNotification{
		Body:    t.Sender + " accepted your friend invite",
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
		Where("(sender = ? OR receiver = ?) AND accepted = 1", username, username).
		Find(&invites).
		Error; err != nil {
		return nil, err
	}

	usernames := GetUsernamesFromInvites(invites, username)

	var usersArray []users.UserData
	if err := db.
		Table("users").
		Where("username IN ?", usernames).
		Find(&usersArray).Error; err != nil {
		return nil, err
	}

	var usersData []users.UserData
	for _, username := range usernames {
		user := getUser(usersArray, username)

		usersData = append(usersData, user)
	}

	return usersData, nil
}

// GetFriendRequests gets friend requests from invites
func GetFriendRequests(db *gorm.DB, username string) (*[]users.UserData, error) {
	var invites []Invite

	if err := db.
		Table("invites").
		Where("receiver = ? AND accepted = 0", username).
		Find(&invites).
		Error; err != nil {
		return nil, err
	}

	usernames := GetUsernamesFromInvites(invites, username)

	var usersData []users.UserData
	if err := db.
		Table("users").
		Where("username IN ?", usernames).
		Find(&usersData).Error; err != nil {
		return nil, err
	}

	return &usersData, nil
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

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("invites").
			Where("(sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)",
				user, username, username, user).
			Delete(&Invite{}).
			Error
	})
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

// helper function to get user form users array
func getUser(usersArray []users.UserData, username string) users.UserData {
	for _, user := range usersArray {
		if user.Username == username {
			return user
		}
	}

	return users.UserData{}
}
