package chats

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/radekkrejcirik01/Koala-backend/pkg/model/connections"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/emotions"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/settings"
	"github.com/radekkrejcirik01/Koala-backend/pkg/model/users"
	"github.com/radekkrejcirik01/Koala-backend/pkg/service"
	"gorm.io/gorm"
)

const MessageType = "message"
const EmotionType = "emotion"

const MessageActivity = "message"
const ReactionActivity = "reaction"

const pageLimit = 20

// rangeLimit caps emotions returned for a time range, like the mood overview
const rangeLimit = 500

var ErrNotFriend = errors.New("user is not your friend")

// previewLimit caps quoted message text in reaction notifications and reply previews
const previewLimit = 150

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
	LowLastReadMessageId  int64  `gorm:"default:0"`
	HighLastReadMessageId int64  `gorm:"default:0"`
	CreatedAt             int64  `gorm:"autoCreateTime"`
}

func (Chat) TableName() string {
	return "chats"
}

type Message struct {
	Id        uint  `gorm:"primary_key;auto_increment;not_null"`
	ChatId    int64 `gorm:"index"`
	SenderId  int64
	Type      string  `gorm:"size:20"`
	Message   string  `gorm:"size:512"`
	Emotion   *string `gorm:"size:64"`
	Reaction  *string `gorm:"size:64"`
	ReplyToId *int64
	Time      int64 `gorm:"autoCreateTime"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageData struct {
	Id       int64      `json:"id"`
	SenderId int64      `json:"senderId"`
	Type     string     `json:"type"`
	Message  string     `json:"message"`
	Emotion  *string    `json:"emotion,omitempty"`
	Reaction *string    `json:"reaction,omitempty"`
	ReplyTo  *ReplyData `json:"replyTo,omitempty"`
	Time     int64      `json:"time"`
}

// ReplyData previews the replied message, only id and deleted are set when it was deleted
type ReplyData struct {
	Id       int64   `json:"id"`
	SenderId int64   `json:"senderId,omitempty"`
	Type     string  `json:"type,omitempty"`
	Message  string  `json:"message,omitempty"`
	Emotion  *string `json:"emotion,omitempty"`
	Deleted  bool    `json:"deleted,omitempty"`
}

type ActivityData struct {
	Type    string      `json:"type"`
	UserId  int64       `json:"userId"`
	TimeMs  int64       `json:"timeMs"`
	Message MessageData `json:"message"`
}

type EmotionsData struct {
	Messages    []MessageData
	UserCount   *int64
	FriendCount *int64
}

type ChatData struct {
	Id           int64          `json:"id"`
	Friend       users.UserData `json:"friend"`
	LastActivity ActivityData   `json:"lastActivity"`
	Unread       bool           `json:"unread"`
}

type SendMessageRequest struct {
	Message   string
	ReplyToId *int64
}

type SendEmotionRequest struct {
	Ids     []int64
	Emotion *string
	Message string
}

type ReactionRequest struct {
	Reaction *string
}

// Socket events data, friendId is the other user of the chat from the receiver's view

type MessageEventData struct {
	FriendId int64       `json:"friendId"`
	Message  MessageData `json:"message"`
}

type MessageDeletedEventData struct {
	FriendId int64 `json:"friendId"`
	Id       int64 `json:"id"`
}

type MessageReactionEventData struct {
	FriendId int64   `json:"friendId"`
	Id       int64   `json:"id"`
	Reaction *string `json:"reaction"`
}

type ChatReadEventData struct {
	FriendId          int64 `json:"friendId"`
	ReaderId          int64 `json:"readerId"`
	LastReadMessageId int64 `json:"lastReadMessageId"`
	LastReadAt        int64 `json:"lastReadAt"`
}

type TypingEventData struct {
	FriendId int64 `json:"friendId"`
	Typing   bool  `json:"typing"`
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

	if err := settings.SetNicknames(db, username, friends); err != nil {
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

// GetMessages gets messages of the chat with a friend, newest first,
// and the id of the last message user read before opening the chat.
// Older messages are paged by lastId, newer ones by firstId, aroundId gets the messages
// around the given one, used to jump to a search result
func GetMessages(db *gorm.DB, username, friendId, lastId, firstId, aroundId string) ([]MessageData, int64, error) {
	user, err := getUser(db, username)
	if err != nil {
		return nil, 0, err
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return nil, 0, err
	}

	lastReadId := chat.HighLastReadMessageId
	if chat.UserLowId == user.Id {
		lastReadId = chat.LowLastReadMessageId
	}

	var messages []Message
	switch {
	case aroundId != "":
		// The message itself is the first of the older ones
		older, err := findMessages(db, int64(chat.Id), "id <= ?", aroundId, "id DESC", pageLimit+1)
		if err != nil {
			return nil, 0, err
		}
		if len(older) == 0 || strconv.FormatInt(int64(older[0].Id), 10) != aroundId {
			return nil, 0, errors.New("message not found")
		}

		newer, err := findMessages(db, int64(chat.Id), "id > ?", aroundId, "id ASC", pageLimit)
		if err != nil {
			return nil, 0, err
		}

		messages = append(reversed(newer), older...)
	case firstId != "":
		newer, err := findMessages(db, int64(chat.Id), "id > ?", firstId, "id ASC", pageLimit)
		if err != nil {
			return nil, 0, err
		}

		messages = reversed(newer)
	case lastId != "":
		messages, err = findMessages(db, int64(chat.Id), "id < ?", lastId, "id DESC", pageLimit)
	default:
		messages, err = findMessages(db, int64(chat.Id), "", nil, "id DESC", pageLimit)
	}
	if err != nil {
		return nil, 0, err
	}

	messagesData, err := toMessagesData(db, messages)
	if err != nil {
		return nil, 0, err
	}

	return messagesData, lastReadId, nil
}

// SearchMessages searches text of messages and emotions in the chat with a friend,
// ignoring case, newest first, paginated by lastId
func SearchMessages(db *gorm.DB, username, friendId, q, lastId string) ([]MessageData, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, errors.New("query is required")
	}

	user, err := getUser(db, username)
	if err != nil {
		return nil, err
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return nil, err
	}

	// The column's collation ignores case
	query := db.
		Table("messages").
		Where("chat_id = ? AND message LIKE ?", chat.Id, "%"+escapeLike(q)+"%")

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

	return toMessagesData(db, messages)
}

// GetEmotions gets emotions shared in the chat with a friend, newest first, senderId filters them by sender.
// The list is paginated by lastId, its first page also has the count of each user's emotions.
// With from it gets all friend's emotions since then, like for the mood overview, only between friends
func GetEmotions(db *gorm.DB, username, friendId, senderId, lastId, from string) (EmotionsData, error) {
	user, err := getUser(db, username)
	if err != nil {
		return EmotionsData{}, err
	}

	friend, err := strconv.ParseInt(friendId, 10, 64)
	if err != nil {
		return EmotionsData{}, err
	}

	if senderId != "" && senderId != strconv.FormatInt(user.Id, 10) && senderId != friendId {
		return EmotionsData{}, errors.New("sender must be you or the friend")
	}

	if from != "" {
		count, err := countFriends(db, user.Username, []int64{friend})
		if err != nil {
			return EmotionsData{}, err
		}
		if count == 0 {
			return EmotionsData{}, ErrNotFriend
		}
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return EmotionsData{}, err
	}

	query := db.
		Table("messages").
		Where("chat_id = ? AND type = ?", chat.Id, EmotionType)

	if senderId != "" {
		query = query.Where("sender_id = ?", senderId)
	}

	limit := pageLimit
	if from != "" {
		query = query.Where("time >= ?", from)
		limit = rangeLimit
	} else if lastId != "" {
		query = query.Where("id < ?", lastId)
	}

	var messages []Message
	if err := query.
		Order("id DESC").
		Limit(limit).
		Find(&messages).
		Error; err != nil {
		return EmotionsData{}, err
	}

	messagesData, err := toMessagesData(db, messages)
	if err != nil {
		return EmotionsData{}, err
	}

	data := EmotionsData{Messages: messagesData}

	if from != "" || lastId != "" {
		return data, nil
	}

	var counts []struct {
		SenderId int64
		Count    int64
	}
	if err := db.
		Table("messages").
		Select("sender_id, COUNT(*) AS count").
		Where("chat_id = ? AND type = ?", chat.Id, EmotionType).
		Group("sender_id").
		Find(&counts).
		Error; err != nil {
		return EmotionsData{}, err
	}

	var userCount, friendCount int64
	for _, c := range counts {
		if c.SenderId == user.Id {
			userCount = c.Count
		} else if c.SenderId == friend {
			friendCount = c.Count
		}
	}
	data.UserCount = &userCount
	data.FriendCount = &friendCount

	return data, nil
}

// GetSettings gets user's private settings of the chat with a friend
func GetSettings(db *gorm.DB, username, friendId string) (settings.SettingsData, error) {
	user, err := getUser(db, username)
	if err != nil {
		return settings.SettingsData{}, err
	}

	id, err := strconv.ParseInt(friendId, 10, 64)
	if err != nil {
		return settings.SettingsData{}, err
	}

	return settings.Get(db, user.Id, id)
}

// UpdateSettings updates user's private settings of the chat with a friend,
// missing fields stay unchanged, null or an empty value removes the setting
func UpdateSettings(db *gorm.DB, username, friendId string, t *settings.UpdateRequest) error {
	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(friendId, 10, 64)
	if err != nil {
		return err
	}

	count, err := countFriends(db, user.Username, []int64{id})
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFriend
	}

	setting := settings.Setting{UserId: user.Id, FriendId: id}
	var columns []string

	if t.Nickname.Set {
		if setting.Nickname, err = settingValue(t.Nickname.Value, settings.NicknameLimit, "nickname"); err != nil {
			return err
		}
		columns = append(columns, "nickname")
	}
	if t.Theme.Set {
		if setting.Theme, err = settingValue(t.Theme.Value, settings.ThemeLimit, "theme"); err != nil {
			return err
		}
		columns = append(columns, "theme")
	}

	if len(columns) == 0 {
		return nil
	}

	return settings.Save(db, setting, columns)
}

// ReadChat moves user's read position to the latest activity and last message of the chat
func ReadChat(db *gorm.DB, username, friendId string) error {
	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	chat, err := getChat(db, user.Id, friendId)
	if err != nil || chat.Id == 0 {
		return err
	}

	prefix := "high"
	if chat.UserLowId == user.Id {
		prefix = "low"
	}

	if err := db.
		Table("chats").
		Where("id = ?", chat.Id).
		Updates(map[string]interface{}{
			prefix + "_last_read_at":         gorm.Expr("last_activity_at"),
			prefix + "_last_read_message_id": gorm.Expr("last_message_id"),
		}).
		Error; err != nil {
		return err
	}

	connections.Send(db, toBoth(user.Id, friendIdOf(chat, user.Id), connections.ChatReadEvent,
		func(friendId int64) interface{} {
			return ChatReadEventData{
				FriendId:          friendId,
				ReaderId:          user.Id,
				LastReadMessageId: chat.LastMessageId,
				LastReadAt:        chat.LastActivityAt,
			}
		}))

	return nil
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
	var replyTo *ReplyData
	err = db.Transaction(func(tx *gorm.DB) error {
		chat, err := getOrCreateChat(tx, user, receiverId)
		if err != nil {
			return err
		}

		if t.ReplyToId != nil {
			// The replied message must be from the same chat
			var replied Message
			if err := tx.
				Table("messages").
				Where("id = ? AND chat_id = ?", *t.ReplyToId, chat.Id).
				First(&replied).
				Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("replied message not found")
				}
				return err
			}
			reply := toReplyData(replied)
			replyTo = &reply
		}

		message = Message{
			ChatId:    int64(chat.Id),
			SenderId:  user.Id,
			Type:      MessageType,
			Message:   t.Message,
			ReplyToId: t.ReplyToId,
		}

		return createMessage(tx, &message)
	})
	if err != nil {
		return MessageData{}, err
	}

	messageData := toMessageData(message)
	messageData.ReplyTo = replyTo
	connections.Send(db, toBoth(user.Id, receiverId, connections.MessageNewEvent,
		func(friendId int64) interface{} {
			return MessageEventData{FriendId: friendId, Message: messageData}
		}))

	if err := notify(db, user, []int64{receiverId}, "", t.Message); err != nil {
		return MessageData{}, err
	}

	return messageData, nil
}

// SendEmotion sends an emotion to the chats with friends and saves it to emotions history,
// without ids the emotion is only logged to the history
func SendEmotion(db *gorm.DB, username string, t *SendEmotionRequest) error {
	t.Message = strings.TrimSpace(t.Message)
	if t.Message == "" && (t.Emotion == nil || *t.Emotion == "") {
		return errors.New("emotion or message is required")
	}

	user, err := getUser(db, username)
	if err != nil {
		return err
	}

	ids := uniqueIds(t.Ids)

	var messages []Message
	err = db.Transaction(func(tx *gorm.DB) error {
		emotion := emotions.Emotion{
			UserId:  user.Id,
			Emotion: t.Emotion,
			Message: t.Message,
		}
		if err := tx.Table("emotions").Create(&emotion).Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			return nil
		}

		recipients := make([]emotions.Recipient, 0, len(ids))
		for _, id := range ids {
			recipients = append(recipients, emotions.Recipient{
				EmotionId: int64(emotion.Id),
				UserId:    id,
			})
		}
		if err := tx.Table("emotion_recipients").Create(&recipients).Error; err != nil {
			return err
		}

		chats, err := getOrCreateChats(tx, user, ids)
		if err != nil {
			return err
		}

		messages = make([]Message, 0, len(chats))
		for _, chat := range chats {
			messages = append(messages, Message{
				ChatId:   int64(chat.Id),
				SenderId: user.Id,
				Type:     EmotionType,
				Message:  t.Message,
				Emotion:  t.Emotion,
			})
		}
		if err := tx.Table("messages").Create(&messages).Error; err != nil {
			return err
		}

		for _, message := range messages {
			if err := setLastMessage(tx, message); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	// Messages are in the order of ids
	var deliveries []connections.Delivery
	for i, message := range messages {
		messageData := toMessageData(message)
		deliveries = append(deliveries, toBoth(user.Id, ids[i], connections.MessageNewEvent,
			func(friendId int64) interface{} {
				return MessageEventData{FriendId: friendId, Message: messageData}
			})...)
	}
	connections.Send(db, deliveries)

	body := t.Message
	if t.Emotion != nil {
		body = strings.TrimSpace(body + " " + *t.Emotion)
	}

	return notify(db, user, ids, " is feeling", body)
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

	connections.Send(db, toBoth(user.Id, message.SenderId, connections.MessageReactionEvent,
		func(friendId int64) interface{} {
			return MessageReactionEventData{FriendId: friendId, Id: int64(message.Id), Reaction: reaction}
		}))

	if reaction == nil {
		return nil
	}

	body := "Reacted " + *reaction + " to \"" + messagePreview(message) + "\""

	// A reaction on an emotion is a support
	if message.Type == EmotionType {
		body = "Sent support 🫶 to \"" + messagePreview(message) + "\""
	}

	return notify(db, user, []int64{message.SenderId}, "", body)
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

	var chat Chat
	if err := db.
		Table("chats").
		Where("id = ?", message.ChatId).
		First(&chat).
		Error; err != nil {
		return err
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
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
	}); err != nil {
		return err
	}

	connections.Send(db, toBoth(user.Id, friendIdOf(chat, user.Id), connections.MessageDeletedEvent,
		func(friendId int64) interface{} {
			return MessageDeletedEventData{FriendId: friendId, Id: int64(message.Id)}
		}))

	return nil
}

// SendTyping sends user's typing state to the friend
func SendTyping(db *gorm.DB, user users.UserData, friendId int64, typing bool) error {
	count, err := countFriends(db, user.Username, []int64{friendId})
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFriend
	}

	connections.Send(db, []connections.Delivery{{
		UserId: friendId,
		Event: connections.Event{
			Type: connections.TypingEvent,
			Data: TypingEventData{FriendId: user.Id, Typing: typing},
		},
	}})

	return nil
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
	chats, err := getOrCreateChats(tx, user, []int64{friendId})
	if err != nil {
		return Chat{}, err
	}

	return chats[0], nil
}

// getOrCreateChats gets the chats between user and friends in the order of friendIds,
// new chats can be created only between friends
func getOrCreateChats(tx *gorm.DB, user users.UserData, friendIds []int64) ([]Chat, error) {
	for _, id := range friendIds {
		if id == user.Id {
			return nil, errors.New("cannot send a message to yourself")
		}
	}

	var existing []Chat
	if err := tx.
		Table("chats").
		Where("(user_low_id = ? AND user_high_id IN ?) OR (user_high_id = ? AND user_low_id IN ?)",
			user.Id, friendIds, user.Id, friendIds).
		Find(&existing).
		Error; err != nil {
		return nil, err
	}

	chatsMap := make(map[int64]Chat, len(friendIds))
	for _, chat := range existing {
		chatsMap[friendIdOf(chat, user.Id)] = chat
	}

	var missingIds []int64
	for _, id := range friendIds {
		if _, ok := chatsMap[id]; !ok {
			missingIds = append(missingIds, id)
		}
	}

	if len(missingIds) > 0 {
		count, err := countFriends(tx, user.Username, missingIds)
		if err != nil {
			return nil, err
		}
		if count != int64(len(missingIds)) {
			return nil, ErrNotFriend
		}

		newChats := make([]Chat, 0, len(missingIds))
		for _, id := range missingIds {
			low, high := sortIds(user.Id, id)
			newChats = append(newChats, Chat{
				UserLowId:  low,
				UserHighId: high,
			})
		}
		if err := tx.Table("chats").Create(&newChats).Error; err != nil {
			return nil, err
		}

		for _, chat := range newChats {
			chatsMap[friendIdOf(chat, user.Id)] = chat
		}
	}

	chats := make([]Chat, 0, len(friendIds))
	for _, id := range friendIds {
		chats = append(chats, chatsMap[id])
	}

	return chats, nil
}

// countFriends counts how many of the given users are user's friends
func countFriends(db *gorm.DB, username string, friendIds []int64) (int64, error) {
	var count int64
	err := db.
		Table("users").
		Joins(`JOIN invites ON invites.accepted = 1 AND (
			(invites.sender = ? AND invites.receiver = users.username) OR
			(invites.sender = users.username AND invites.receiver = ?))`,
			username, username).
		Where("users.id IN ?", friendIds).
		Distinct("users.id").
		Count(&count).
		Error

	return count, err
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

	return setLastMessage(tx, *message)
}

// setLastMessage sets the message as chat's last message and activity in one update
func setLastMessage(tx *gorm.DB, message Message) error {
	// The sender has read the chat up to their own message
	values := activityValues(message.SenderId, MessageActivity, int64(message.Id), message.Id)
	values["last_message_id"] = gorm.Expr("GREATEST(last_message_id, ?)", message.Id)

	return tx.
		Table("chats").
		Where("id = ?", message.ChatId).
		Updates(values).
		Error
}

// setActivity sets chat's latest activity, the user who made it has read the chat
func setActivity(tx *gorm.DB, chatId, userId int64, activityType string, messageId int64) error {
	return tx.
		Table("chats").
		Where("id = ?", chatId).
		Updates(activityValues(userId, activityType, messageId, gorm.Expr("last_message_id"))).
		Error
}

// activityValues sets the latest activity, readMessageId is the last message the user who made it has read
func activityValues(userId int64, activityType string, messageId int64, readMessageId interface{}) map[string]interface{} {
	now := time.Now().UnixMilli()

	return map[string]interface{}{
		"last_activity_at":          now,
		"last_activity_type":        activityType,
		"last_activity_user_id":     userId,
		"last_activity_message_id":  messageId,
		"low_last_read_at":          gorm.Expr("IF(user_low_id = ?, ?, low_last_read_at)", userId, now),
		"high_last_read_at":         gorm.Expr("IF(user_high_id = ?, ?, high_last_read_at)", userId, now),
		"low_last_read_message_id":  gorm.Expr("IF(user_low_id = ?, ?, low_last_read_message_id)", userId, readMessageId),
		"high_last_read_message_id": gorm.Expr("IF(user_high_id = ?, ?, high_last_read_message_id)", userId, readMessageId),
	}
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

// notify sends the notification to receivers' devices, the title starts with sender's name
// or the nickname the receiver gave to the sender, each title is sent as its own notification
func notify(db *gorm.DB, sender users.UserData, receiverIds []int64, titleSuffix, body string) error {
	var devices []struct {
		UserId      int64
		DeviceToken string
	}
	if err := db.
		Table("devices").
		Select("user_id, device_token").
		Distinct("user_id, device_token").
		Where("user_id IN ?", receiverIds).
		Find(&devices).
		Error; err != nil {
		return err
	}

	if len(devices) == 0 {
		return nil
	}

	nicknames, err := settings.GetSenderNicknames(db, sender.Id, receiverIds)
	if err != nil {
		return err
	}

	var titles []string
	titleTokens := make(map[string][]string)
	for _, device := range devices {
		name := sender.Name
		if nickname, ok := nicknames[device.UserId]; ok {
			name = nickname
		}

		title := name + titleSuffix
		if _, ok := titleTokens[title]; !ok {
			titles = append(titles, title)
		}
		titleTokens[title] = append(titleTokens[title], device.DeviceToken)
	}

	for _, title := range titles {
		fcmNotification := service.FcmNotification{
			Data: map[string]string{
				"type":     "chat",
				"friendId": strconv.FormatInt(sender.Id, 10),
			},
			Title:   title,
			Body:    body,
			Sound:   "default",
			Devices: titleTokens[title],
		}

		if err := service.SendNotification(&fcmNotification); err != nil {
			return err
		}
	}

	return nil
}

// findMessages gets up to limit messages of the chat matching the condition in the order
func findMessages(db *gorm.DB, chatId int64, condition string, arg interface{}, order string, limit int) ([]Message, error) {
	query := db.
		Table("messages").
		Where("chat_id = ?", chatId)

	if condition != "" {
		query = query.Where(condition, arg)
	}

	var messages []Message
	err := query.
		Order(order).
		Limit(limit).
		Find(&messages).
		Error

	return messages, err
}

func reversed(messages []Message) []Message {
	result := make([]Message, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		result = append(result, messages[i])
	}

	return result
}

// escapeLike escapes LIKE wildcards so the text is matched literally
func escapeLike(text string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(text)
}

// settingValue trims the value, an empty value removes the setting
func settingValue(value *string, limit int, name string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > limit {
		return nil, errors.New(name + " can have at most " + strconv.Itoa(limit) + " characters")
	}

	return &trimmed, nil
}

func messagePreview(message Message) string {
	preview := truncate(message.Message, previewLimit)
	if message.Emotion != nil {
		return strings.TrimSpace(preview + " " + *message.Emotion)
	}

	return preview
}

// truncate cuts text to limit characters, keeps multi-byte characters like emoji whole
func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return strings.TrimSpace(string(runes[:limit])) + "…"
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

// toMessagesData converts the messages with previews of the replied messages, loaded in one query
func toMessagesData(db *gorm.DB, messages []Message) ([]MessageData, error) {
	var replyIds []int64
	for _, message := range messages {
		if message.ReplyToId != nil {
			replyIds = append(replyIds, *message.ReplyToId)
		}
	}

	repliedMap := make(map[int64]Message)
	if len(replyIds) > 0 {
		var replied []Message
		if err := db.
			Table("messages").
			Where("id IN ?", uniqueIds(replyIds)).
			Find(&replied).
			Error; err != nil {
			return nil, err
		}

		for _, message := range replied {
			repliedMap[int64(message.Id)] = message
		}
	}

	var messagesData []MessageData
	for _, message := range messages {
		messageData := toMessageData(message)
		if message.ReplyToId != nil {
			reply := ReplyData{Id: *message.ReplyToId, Deleted: true}
			if replied, ok := repliedMap[*message.ReplyToId]; ok {
				reply = toReplyData(replied)
			}
			messageData.ReplyTo = &reply
		}
		messagesData = append(messagesData, messageData)
	}

	return messagesData, nil
}

func toReplyData(message Message) ReplyData {
	return ReplyData{
		Id:       int64(message.Id),
		SenderId: message.SenderId,
		Type:     message.Type,
		Message:  truncate(message.Message, previewLimit),
		Emotion:  message.Emotion,
	}
}

// toBoth creates the event for both users of the chat, each with the other user as friendId,
// user's event also reaches user's other devices
func toBoth(userId, friendId int64, eventType string, data func(friendId int64) interface{}) []connections.Delivery {
	return []connections.Delivery{
		{UserId: friendId, Event: connections.Event{Type: eventType, Data: data(userId)}},
		{UserId: userId, Event: connections.Event{Type: eventType, Data: data(friendId)}},
	}
}

func uniqueIds(ids []int64) []int64 {
	seen := make(map[int64]bool, len(ids))
	var unique []int64
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
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
