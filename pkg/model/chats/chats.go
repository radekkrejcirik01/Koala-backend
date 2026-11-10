package chats

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const MessageType = "message"
const EmotionType = "emotion"

const MessageActivity = "message"
const ReactionActivity = "reaction"

const pageLimit = 20

type Chat struct {
	Id                    uint   `gorm:"primary_key;auto_increment;not_null"`
	UserLowId             int64  `gorm:"uniqueIndex:idx_chats_users"`
	UserHighId            int64  `gorm:"uniqueIndex:idx_chats_users;index"`
	LastMessageId         int64  `gorm:"default:0"`
	LastActivityAt        int64  `gorm:"default:0"`
	LastActivityType      string `gorm:"size:20"`
	LastActivityUserId    int64  `gorm:"default:0"`
	LastActivityMessageId int64  `gorm:"default:0"`
	LowLastReadAt         int64  `gorm:"default:0"`
	HighLastReadAt        int64  `gorm:"default:0"`
	CreatedAt             int64  `gorm:"autoCreateTime"`
}

func (Chat) TableName() string {
	return "chats"
}

type Message struct {
	Id       uint  `gorm:"primary_key;auto_increment;not_null"`
	ChatId   int64 `gorm:"index"`
	SenderId int64
	Type     string  `gorm:"size:20"`
	Message  string  `gorm:"size:512"`
	Emotion  *string `gorm:"size:64"`
	Reaction *string `gorm:"size:64"`
	Time     int64   `gorm:"autoCreateTime"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageData struct {
	Id       int64   `json:"id"`
	SenderId int64   `json:"senderId"`
	Type     string  `json:"type"`
	Message  string  `json:"message"`
	Emotion  *string `json:"emotion,omitempty"`
	Reaction *string `json:"reaction,omitempty"`
	Time     int64   `json:"time"`
}

type ActivityData struct {
	Type    string      `json:"type"`
	UserId  int64       `json:"userId"`
	TimeMs  int64       `json:"timeMs"`
	Message MessageData `json:"message"`
}

type ChatData struct {
	Id           int64          `json:"id"`
	Friend       users.UserData `json:"friend"`
	LastActivity ActivityData   `json:"lastActivity"`
	Unread       bool           `json:"unread"`
}

type SendMessageRequest struct {
	Message string
}

type SendEmotionRequest struct {
	Ids     []int64
	Emotion *string
	Message string
}

type ReactionRequest struct {
	Reaction *string
}

// GetChats gets user's chats ordered by the latest activity, a new message or a reaction
func GetChats(db *gorm.DB, username, lastTimeMs, lastId string) ([]ChatData, error) {
	user, err := getUser(db, username)
	if err != nil {
		return nil, err
	}

	query := db.
		Table("chats").
		Where("(user_low_id = ? OR user_high_id = ?) AND last_message_id > 0", user.Id, user.Id)

	if lastTimeMs != "" && lastId != "" {
		query = query.Where("(last_activity_at < ? OR (last_activity_at = ? AND id < ?))",
			lastTimeMs, lastTimeMs, lastId)
	} else if lastTimeMs != "" {
		query = query.Where("last_activity_at < ?", lastTimeMs)
	}

	var chats []Chat
	if err := query.
		Order("last_activity_at DESC, id DESC").
		Limit(pageLimit).
		Find(&chats).
		Error; err != nil {
		return nil, err
	}

	if len(chats) == 0 {
		return nil, nil
	}

	var friendIds, messageIds []int64
	for _, chat := range chats {
		friendIds = append(friendIds, friendIdOf(chat, user.Id))
		messageIds = append(messageIds, chat.LastActivityMessageId)
	}

	var friends []users.UserData
	if err := db.
		Table("users").
		Select("id, username, name, profile_photo").
		Where("id IN ?", friendIds).
		Find(&friends).
		Error; err != nil {
		return nil, err
	}

	var messages []Message
	if err := db.
		Table("messages").
		Where("id IN ?", messageIds).
		Find(&messages).
		Error; err != nil {
		return nil, err
	}

	friendsMap := make(map[int64]users.UserData)
	for _, friend := range friends {
		friendsMap[friend.Id] = friend
	}

	messagesMap := make(map[int64]Message)
	for _, message := range messages {
		messagesMap[int64(message.Id)] = message
	}

	var chatsData []ChatData
	for _, chat := range chats {
		chatsData = append(chatsData, ChatData{
			Id:     int64(chat.Id),
			Friend: friendsMap[friendIdOf(chat, user.Id)],
			LastActivity: ActivityData{
				Type:    chat.LastActivityType,
				UserId:  chat.LastActivityUserId,
				TimeMs:  chat.LastActivityAt,
				Message: toMessageData(messagesMap[chat.LastActivityMessageId]),
			},
			Unread: isUnread(chat, user.Id),
		})
	}

	return chatsData, nil
}

// GetUnreadChats counts chats where the latest activity is from the friend and wasn't read yet
func GetUnreadChats(db *gorm.DB, username string) (int64, error) {
	user, err := getUser(db, username)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.
		Table("chats").
		Where(`last_message_id > 0 AND last_activity_user_id <> ? AND (
			(user_low_id = ? AND last_activity_at > low_last_read_at) OR
			(user_high_id = ? AND last_activity_at > high_last_read_at))`,
			user.Id, user.Id, user.Id).
		Count(&count).
		Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetMessages gets messages of the chat with a friend, newest first
func GetMessages(db *gorm.DB, username, friendId, lastId string) ([]MessageData, error) {
	user, err := getUser(db, username)
	if err != nil {
		return nil, err
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return nil, err
	}

	query := db.
		Table("messages").
		Where("chat_id = ?", chat.Id)

	if lastId != "" {
		query = query.Where("id < ?", lastId)
	}

	var messages []Message
	if err := query.
		Order("id DESC").
		Limit(pageLimit).
		Find(&messages).
		Error; err != nil {
		return nil, err
	}

	var messagesData []MessageData
	for _, message := range messages {
		messagesData = append(messagesData, toMessageData(message))
	}

	return messagesData, nil
}

// ReadChat moves user's read position to the latest activity of the chat
func ReadChat(db *gorm.DB, username, friendId string) error {
	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return err
	}

	column := "high_last_read_at"
	if chat.UserLowId == user.Id {
		column = "low_last_read_at"
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("chats").
			Where("id = ?", chat.Id).
			Update(column, gorm.Expr("last_activity_at")).
			Error
	})
}

// SendMessage sends a text message to the chat with a friend
func SendMessage(db *gorm.DB, username, friendId string, t *SendMessageRequest) (MessageData, error) {
	t.Message = strings.TrimSpace(t.Message)
	if t.Message == "" {
		return MessageData{}, errors.New("message is required")
	}

	user, err := getUser(db, username)
	if err != nil {
		return MessageData{}, err
	}

	receiverId, err := strconv.ParseInt(friendId, 10, 64)
	if err != nil {
		return MessageData{}, err
	}

	var message Message
	err = db.Transaction(func(tx *gorm.DB) error {
		chat, err := getOrCreateChat(tx, user, receiverId)
		if err != nil {
			return err
		}

		message = Message{
			ChatId:   int64(chat.Id),
			SenderId: user.Id,
			Type:     MessageType,
			Message:  t.Message,
		}

		return createMessage(tx, &message)
	})
	if err != nil {
		return MessageData{}, err
	}

	if err := notify(db, user, []int64{receiverId}, user.Name, t.Message); err != nil {
		return MessageData{}, err
	}

	return toMessageData(message), nil
}

// SendEmotion sends an emotion to the chats with friends and saves it to emotions history
func SendEmotion(db *gorm.DB, username string, t *SendEmotionRequest) error {
	t.Message = strings.TrimSpace(t.Message)
	if len(t.Ids) == 0 {
		return errors.New("ids are required")
	}
	if t.Message == "" && (t.Emotion == nil || *t.Emotion == "") {
		return errors.New("emotion or message is required")
	}

	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		emotion := emotions.Emotion{
			UserId:  user.Id,
			Emotion: t.Emotion,
			Message: t.Message,
		}
		if err := tx.Table("emotions").Create(&emotion).Error; err != nil {
			return err
		}

		for _, id := range t.Ids {
			chat, err := getOrCreateChat(tx, user, id)
			if err != nil {
				return err
			}

			message := Message{
				ChatId:   int64(chat.Id),
				SenderId: user.Id,
				Type:     EmotionType,
				Message:  t.Message,
				Emotion:  t.Emotion,
			}
			if err := createMessage(tx, &message); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	body := t.Message
	if t.Emotion != nil {
		body = strings.TrimSpace(body + " " + *t.Emotion)
	}

	return notify(db, user, t.Ids, user.Name+" is feeling", body)
}

// UpdateReaction sets or removes a reaction on the friend's message
func UpdateReaction(db *gorm.DB, username, id string, reaction *string) error {
	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	message, err := getFriendMessage(db, user.Id, id)
	if err != nil {
		return err
	}

	// Nothing changes, avoids repeated activity and notification
	if sameReaction(message.Reaction, reaction) {
		return nil
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Table("messages").
			Where("id = ?", message.Id).
			Update("reaction", reaction).
			Error; err != nil {
			return err
		}

		if reaction == nil {
			return revertActivity(tx, message.ChatId, int64(message.Id))
		}

		return setActivity(tx, message.ChatId, user.Id, ReactionActivity, int64(message.Id))
	}); err != nil {
		return err
	}

	if reaction == nil {
		return nil
	}

	body := "Reacted " + *reaction + " to \"" + messagePreview(message) + "\""

	// A reaction on an emotion is a support
	if message.Type == EmotionType {
		body = "Sent support 🫶 to \"" + messagePreview(message) + "\""
	}

	return notify(db, user, []int64{message.SenderId}, user.Name, body)
}

// DeleteMessage deletes user's own message and moves chat's last message and activity back if needed
func DeleteMessage(db *gorm.DB, username, id string) error {
	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	var message Message
	if err := db.
		Table("messages").
		Where("id = ? AND sender_id = ?", id, user.Id).
		First(&message).
		Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Table("messages").
			Where("id = ?", message.Id).
			Delete(&Message{}).
			Error; err != nil {
			return err
		}

		var lastMessageId int64
		if err := tx.
			Table("messages").
			Select("COALESCE(MAX(id), 0)").
			Where("chat_id = ?", message.ChatId).
			Find(&lastMessageId).
			Error; err != nil {
			return err
		}

		if err := tx.
			Table("chats").
			Where("id = ? AND last_message_id = ?", message.ChatId, message.Id).
			Update("last_message_id", lastMessageId).
			Error; err != nil {
			return err
		}

		return revertActivity(tx, message.ChatId, int64(message.Id))
	})
}

// DeleteUserChats deletes all chats of the user with their messages
func DeleteUserChats(tx *gorm.DB, userId int64) error {
	chatIds := tx.
		Table("chats").
		Select("id").
		Where("user_low_id = ? OR user_high_id = ?", userId, userId)

	if err := tx.
		Table("messages").
		Where("chat_id IN (?)", chatIds).
		Delete(&Message{}).
		Error; err != nil {
		return err
	}

	return tx.
		Table("chats").
		Where("user_low_id = ? OR user_high_id = ?", userId, userId).
		Delete(&Chat{}).
		Error
}

func getUser(db *gorm.DB, username string) (users.UserData, error) {
	user, err := users.GetUser(db, username)
	if err != nil {
		return users.UserData{}, err
	}
	if user.Id == 0 {
		return users.UserData{}, errors.New("user not found")
	}

	return user, nil
}

// getChat gets the chat between user and friend, chat Id is 0 when it doesn't exist
func getChat(db *gorm.DB, userId int64, friendId string) (Chat, error) {
	id, err := strconv.ParseInt(friendId, 10, 64)
	if err != nil {
		return Chat{}, err
	}

	low, high := sortIds(userId, id)

	var chat Chat
	err = db.
		Table("chats").
		Where("user_low_id = ? AND user_high_id = ?", low, high).
		Limit(1).
		Find(&chat).
		Error

	return chat, err
}

// getOrCreateChat gets the chat between user and friend, a new chat can be created only between friends
func getOrCreateChat(tx *gorm.DB, user users.UserData, friendId int64) (Chat, error) {
	if friendId == user.Id {
		return Chat{}, errors.New("cannot send a message to yourself")
	}

	chat, err := getChat(tx, user.Id, strconv.FormatInt(friendId, 10))
	if err != nil || chat.Id != 0 {
		return chat, err
	}

	friends, err := areFriends(tx, user.Username, friendId)
	if err != nil {
		return Chat{}, err
	}
	if !friends {
		return Chat{}, errors.New("user is not your friend")
	}

	low, high := sortIds(user.Id, friendId)
	chat = Chat{
		UserLowId:  low,
		UserHighId: high,
	}

	err = tx.Table("chats").Create(&chat).Error

	return chat, err
}

func areFriends(db *gorm.DB, username string, friendId int64) (bool, error) {
	var friendUsername string
	if err := db.
		Table("users").
		Select("username").
		Where("id = ?", friendId).
		Find(&friendUsername).
		Error; err != nil {
		return false, err
	}

	if friendUsername == "" {
		return false, nil
	}

	var count int64
	if err := db.
		Table("invites").
		Where("((sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)) AND accepted = 1",
			username, friendUsername, friendUsername, username).
		Count(&count).
		Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// getFriendMessage gets a message sent by the friend in one of user's chats
func getFriendMessage(db *gorm.DB, userId int64, id string) (Message, error) {
	var message Message
	if err := db.
		Table("messages").
		Joins("JOIN chats ON chats.id = messages.chat_id").
		Where("messages.id = ? AND messages.sender_id <> ? AND (chats.user_low_id = ? OR chats.user_high_id = ?)",
			id, userId, userId, userId).
		Select("messages.*").
		First(&message).
		Error; err != nil {
		return Message{}, err
	}

	return message, nil
}

// createMessage creates a message and sets it as chat's last message and activity
func createMessage(tx *gorm.DB, message *Message) error {
	if err := tx.Table("messages").Create(message).Error; err != nil {
		return err
	}

	if err := tx.
		Table("chats").
		Where("id = ?", message.ChatId).
		Update("last_message_id", gorm.Expr("GREATEST(last_message_id, ?)", message.Id)).
		Error; err != nil {
		return err
	}

	return setActivity(tx, message.ChatId, message.SenderId, MessageActivity, int64(message.Id))
}

// setActivity sets chat's latest activity, the user who made it has read the chat
func setActivity(tx *gorm.DB, chatId, userId int64, activityType string, messageId int64) error {
	now := time.Now().UnixMilli()

	return tx.
		Table("chats").
		Where("id = ?", chatId).
		Updates(map[string]interface{}{
			"last_activity_at":         now,
			"last_activity_type":       activityType,
			"last_activity_user_id":    userId,
			"last_activity_message_id": messageId,
			"low_last_read_at":         gorm.Expr("IF(user_low_id = ?, ?, low_last_read_at)", userId, now),
			"high_last_read_at":        gorm.Expr("IF(user_high_id = ?, ?, high_last_read_at)", userId, now),
		}).
		Error
}

// revertActivity sets chat's last message as its latest activity when the latest activity
// was on the given message, used when a reaction is removed or a message is deleted
func revertActivity(tx *gorm.DB, chatId, messageId int64) error {
	var chat Chat
	if err := tx.
		Table("chats").
		Where("id = ?", chatId).
		First(&chat).
		Error; err != nil {
		return err
	}

	if chat.LastActivityMessageId != messageId {
		return nil
	}

	var lastMessage Message
	if err := tx.
		Table("messages").
		Where("id = ?", chat.LastMessageId).
		Limit(1).
		Find(&lastMessage).
		Error; err != nil {
		return err
	}

	return tx.
		Table("chats").
		Where("id = ?", chatId).
		Updates(map[string]interface{}{
			"last_activity_at":         lastMessage.Time * 1000,
			"last_activity_type":       MessageActivity,
			"last_activity_user_id":    lastMessage.SenderId,
			"last_activity_message_id": chat.LastMessageId,
		}).
		Error
}

func notify(db *gorm.DB, sender users.UserData, receiverIds []int64, title, body string) error {
	tokens, err := service.GetTokensByUserIds(db, receiverIds)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	fcmNotification := service.FcmNotification{
		Data: map[string]string{
			"type":     "chat",
			"friendId": strconv.FormatInt(sender.Id, 10),
		},
		Title:   title,
		Body:    body,
		Sound:   "default",
		Devices: tokens,
	}

	return service.SendNotification(&fcmNotification)
}

func messagePreview(message Message) string {
	if message.Emotion != nil {
		return strings.TrimSpace(message.Message + " " + *message.Emotion)
	}

	return message.Message
}

func toMessageData(message Message) MessageData {
	return MessageData{
		Id:       int64(message.Id),
		SenderId: message.SenderId,
		Type:     message.Type,
		Message:  message.Message,
		Emotion:  message.Emotion,
		Reaction: message.Reaction,
		Time:     message.Time,
	}
}

func sortIds(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}

	return b, a
}

func friendIdOf(chat Chat, userId int64) int64 {
	if chat.UserLowId == userId {
		return chat.UserHighId
	}

	return chat.UserLowId
}

func isUnread(chat Chat, userId int64) bool {
	lastReadAt := chat.HighLastReadAt
	if chat.UserLowId == userId {
		lastReadAt = chat.LowLastReadAt
	}

	return chat.LastActivityUserId != userId && chat.LastActivityAt > lastReadAt
}

func sameReaction(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}
