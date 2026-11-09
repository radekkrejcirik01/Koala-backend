package chats

import (
	"errors"
	"strconv"
	"strings"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const MessageType = "message"
const EmotionType = "emotion"
const SupportType = "support"

const pageLimit = 20

type Chat struct {
	Id             uint  `gorm:"primary_key;auto_increment;not_null"`
	UserLowId      int64 `gorm:"uniqueIndex:idx_chats_users"`
	UserHighId     int64 `gorm:"uniqueIndex:idx_chats_users;index"`
	LastMessageId  int64 `gorm:"default:0"`
	LowLastReadId  int64 `gorm:"default:0"`
	HighLastReadId int64 `gorm:"default:0"`
	CreatedAt      int64 `gorm:"autoCreateTime"`
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
	ParentId *int64  `gorm:"index"`
	Time     int64   `gorm:"autoCreateTime"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageData struct {
	Id        int64   `json:"id"`
	SenderId  int64   `json:"senderId"`
	Type      string  `json:"type"`
	Message   string  `json:"message"`
	Emotion   *string `json:"emotion,omitempty"`
	Reaction  *string `json:"reaction,omitempty"`
	ParentId  *int64  `json:"parentId,omitempty"`
	Supported *bool   `json:"supported,omitempty"`
	Time      int64   `json:"time"`
}

type ChatData struct {
	Id          int64          `json:"id"`
	Friend      users.UserData `json:"friend"`
	LastMessage MessageData    `json:"lastMessage"`
	Unread      bool           `json:"unread"`
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

// GetChats gets user's chats ordered by the latest message
func GetChats(db *gorm.DB, username, lastId string) ([]ChatData, error) {
	user, err := getUser(db, username)
	if err != nil {
		return nil, err
	}

	query := db.
		Table("chats").
		Where("(user_low_id = ? OR user_high_id = ?) AND last_message_id > 0", user.Id, user.Id)

	if lastId != "" {
		query = query.Where("last_message_id < ?", lastId)
	}

	var chats []Chat
	if err := query.
		Order("last_message_id DESC").
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
		messageIds = append(messageIds, chat.LastMessageId)
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
		lastMessage := messagesMap[chat.LastMessageId]

		chatsData = append(chatsData, ChatData{
			Id:          int64(chat.Id),
			Friend:      friendsMap[friendIdOf(chat, user.Id)],
			LastMessage: toMessageData(lastMessage),
			Unread:      lastMessage.SenderId != user.Id && chat.LastMessageId > lastReadIdOf(chat, user.Id),
		})
	}

	lastMessages := make([]*MessageData, len(chatsData))
	for i := range chatsData {
		lastMessages[i] = &chatsData[i].LastMessage
	}

	if err := setSupported(db, lastMessages); err != nil {
		return nil, err
	}

	return chatsData, nil
}

// GetUnreadChats counts chats where the latest message is from the friend and wasn't read yet
func GetUnreadChats(db *gorm.DB, username string) (int64, error) {
	user, err := getUser(db, username)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.
		Table("chats").
		Joins("JOIN messages ON messages.id = chats.last_message_id").
		Where(`messages.sender_id <> ? AND (
			(chats.user_low_id = ? AND chats.last_message_id > chats.low_last_read_id) OR
			(chats.user_high_id = ? AND chats.last_message_id > chats.high_last_read_id))`,
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

	pointers := make([]*MessageData, len(messagesData))
	for i := range messagesData {
		pointers[i] = &messagesData[i]
	}

	if err := setSupported(db, pointers); err != nil {
		return nil, err
	}

	return messagesData, nil
}

// ReadChat moves user's read position to the latest message of the chat
func ReadChat(db *gorm.DB, username, friendId string) error {
	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return err
	}

	column := "high_last_read_id"
	if chat.UserLowId == user.Id {
		column = "low_last_read_id"
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("chats").
			Where("id = ?", chat.Id).
			Update(column, gorm.Expr("last_message_id")).
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

// SendSupport sends a support message to the friend's emotion
func SendSupport(db *gorm.DB, username, id string) (MessageData, error) {
	user, err := getUser(db, username)
	if err != nil {
		return MessageData{}, err
	}

	emotion, err := getFriendMessage(db, user.Id, id)
	if err != nil {
		return MessageData{}, err
	}

	if emotion.Type != EmotionType {
		return MessageData{}, errors.New("support can be sent only to an emotion")
	}

	emotionId := int64(emotion.Id)
	support := Message{
		ChatId:   emotion.ChatId,
		SenderId: user.Id,
		Type:     SupportType,
		ParentId: &emotionId,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		return createMessage(tx, &support)
	}); err != nil {
		return MessageData{}, err
	}

	body := "Sent support 🫶 to \"" + messagePreview(emotion) + "\""

	if err := notify(db, user, []int64{emotion.SenderId}, user.Name, body); err != nil {
		return MessageData{}, err
	}

	return toMessageData(support), nil
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

	if err := db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table("messages").
			Where("id = ?", message.Id).
			Update("reaction", reaction).
			Error
	}); err != nil {
		return err
	}

	if reaction == nil {
		return nil
	}

	body := "Reacted " + *reaction + " to \"" + messagePreview(message) + "\""

	return notify(db, user, []int64{message.SenderId}, user.Name, body)
}

// DeleteMessage deletes user's own message and moves chat's last message back if needed
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

		return tx.
			Table("chats").
			Where("id = ? AND last_message_id = ?", message.ChatId, message.Id).
			Update("last_message_id", lastMessageId).
			Error
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

// createMessage creates a message and sets it as chat's last message
func createMessage(tx *gorm.DB, message *Message) error {
	if err := tx.Table("messages").Create(message).Error; err != nil {
		return err
	}

	return tx.
		Table("chats").
		Where("id = ?", message.ChatId).
		Update("last_message_id", gorm.Expr("GREATEST(last_message_id, ?)", message.Id)).
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

// setSupported marks emotion messages which have at least one support
func setSupported(db *gorm.DB, messages []*MessageData) error {
	var emotionIds []int64
	for _, message := range messages {
		if message.Type == EmotionType {
			emotionIds = append(emotionIds, message.Id)
		}
	}

	if len(emotionIds) == 0 {
		return nil
	}

	var supportedIds []int64
	if err := db.
		Table("messages").
		Distinct("parent_id").
		Where("type = ? AND parent_id IN ?", SupportType, emotionIds).
		Find(&supportedIds).
		Error; err != nil {
		return err
	}

	supported := make(map[int64]bool)
	for _, id := range supportedIds {
		supported[id] = true
	}

	for _, message := range messages {
		if message.Type == EmotionType {
			isSupported := supported[message.Id]
			message.Supported = &isSupported
		}
	}

	return nil
}

func messagePreview(message Message) string {
	switch message.Type {
	case SupportType:
		return "Sent support 🫶"
	case EmotionType:
		if message.Emotion != nil {
			return strings.TrimSpace(message.Message + " " + *message.Emotion)
		}
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
		ParentId: message.ParentId,
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

func lastReadIdOf(chat Chat, userId int64) int64 {
	if chat.UserLowId == userId {
		return chat.LowLastReadId
	}

	return chat.HighLastReadId
}
