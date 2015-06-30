package tgbot

import "encoding/json"

// User ...
type User struct {
	ID        int     `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Username  *string `json:"username,omitempty"`
}

// GroupChat ...
type GroupChat struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// UserGroup ..
type UserGroup struct { // For Message type!
	ID        int     `json:"id"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Username  *string `json:"username,omitempty"`
	Title     *string `json:"title,omitempty"`
}

// Message ...
type Message struct {
	ID                  int          `json:"message_id"`
	From                User         `json:"from"`
	Date                int          `json:"date"`
	Chat                UserGroup    `json:"chat"`
	ForwardFrom         *User        `json:"forward_from,omitempty"`
	ForwardDate         *int         `json:"forward_date,omitempty"`
	ReplyToMessage      *Message     `json:"reply_to_message,omitempty"`
	Text                *string      `json:"text,omitempty"`
	Audio               *Audio       `json:"audio,omitempty"`
	Document            *Document    `json:"document,omitempty"`
	Photo               *[]PhotoSize `json:"photo,omitempty"`
	Sticker             *Sticker     `json:"sticker,omitempty"`
	Video               *Video       `json:"video,omitempty"`
	Location            *Location    `json:"location,omitempty"`
	NewChatParticipant  *User        `json:"new_chat_participant,omitempty"`
	LeftChatParticipant *User        `json:"left_chat_participant,omitempty"`
	NewChatTitle        *string      `json:"new_chat_title,omitempty"`
	NewChatPhoto        *string      `json:"new_chat_photo,omitempty"`
	DeleteChatPhoto     *bool        `json:"delete_chat_photo,omitempty"`
	GroupChatCreated    *bool        `json:"group_chat_created,omitempty"`
}

// PhotoSize ...
type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize *int   `json:"file_size,omitempty"`
}

// Audio ..
type Audio struct {
	FileID   string  `json:"file_id"`
	Duration int     `json:"duration"`
	MimeType *string `json:"mime_type,omitempty"`
	FileSize *int    `json:"file_size,omitempty"`
}

// Document ...
type Document struct {
	FileID   string    `json:"file_id"`
	Thumb    PhotoSize `json:"thumb"`
	FileName *string   `json:"file_name,omitempty"`
	MimeType *string   `json:"mime_type,omitempty"`
	FileSize *int      `json:"file_size,omitempty"`
}

// Sticker ...
type Sticker struct {
	FileID   string    `json:"file_id"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Thumb    PhotoSize `json:"thumb"` // .webp or .jpg format
	FileSize *int      `json:"file_size,omitempty"`
}

// Video ...
type Video struct {
	FileID   string    `json:"file_id"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Duration int       `json:"duration"`
	Thumb    PhotoSize `json:"thumb"`
	MimeType *string   `json:"mime_type,omitempty"`
	FileSize *int      `json:"file_size,omitempty"`
	Caption  *string   `json:"caption,omitempty"`
}

// Contact ...
type Contact struct {
	PhoneNumber string  `json:"phone_number"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name,omitempty"`
	UserID      *string `json:"user_id,omitempty"`
}

// Location ...
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// UserPhotos ...
type UserPhotos [][]PhotoSize

// KeyboardLayout ...
type KeyboardLayout [][]string

// UserProfilePhotos ...
type UserProfilePhotos struct {
	TotalCount int        `json:"total_count"`
	Photos     UserPhotos `json:"photos"`
}

// ReplyKeyboardMarkup ...
type ReplyKeyboardMarkup struct {
	Keyboard        KeyboardLayout `json:"keyboard"`
	ResizeKeyboard  bool           `json:"resize_keyboard,omitempty"`   // Default false
	OneTimeKeyboard bool           `json:"one_time_keyboard,omitempty"` // Default false
	Selective       bool           `json:"selective,omitempty"`
}

// ImplementReplyMarkup ...
func (rkm ReplyKeyboardMarkup) ImplementReplyMarkup() {}

// ReplyKeyboardHide ...
type ReplyKeyboardHide struct {
	HideKeyboard bool `json:"hide_keyboard"` // always true!
	Selective    bool `json:"selective,omitempty"`
}

// ImplementReplyMarkup ...
func (rkh ReplyKeyboardHide) ImplementReplyMarkup() {}

// ForceReply ...
type ForceReply struct {
	Force     bool `json:"force_reply"` // always true!
	Selective bool `json:"selective,omitempty"`
}

// ImplementReplyMarkup ...
func (fr ForceReply) ImplementReplyMarkup() {}

// ReplyMarkupInt ...
type ReplyMarkupInt interface {
	ImplementReplyMarkup()
}

// Result messages, this is what we receive from GET params

// ResultBase ...
type ResultBase struct {
	Ok          bool    `json:"ok"`
	ErrorCode   *int    `json:"error_code,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ResultWithMessage ...
type ResultWithMessage struct {
	ResultBase
	Result *Message `json:"result,omitempty"`
}

// MessageWithUpdateID ...
type MessageWithUpdateID struct {
	Msg      Message `json:"message"`
	UpdateID int     `json:"update_id"`
}

// ResultGetUpdates ...
type ResultGetUpdates struct {
	ResultBase
	Result []MessageWithUpdateID `json:"result"`
}

// ResultGetUser ...
type ResultGetUser struct {
	ResultBase
	Result User `json:"result,omitempty"`
}

// QuerySendMessage ...
type QuerySendMessage struct {
	ChatID                int             `json:"chat_id"`
	Text                  string          `json:"text"`
	DisableWebPagePreview *bool           `json:"disable_web_page_preview,omitempty"`
	ReplyToMessageID      *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup           *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// ForwardMessageQuery ...
type ForwardMessageQuery struct {
	ChatID     int `json:"chat_id"`
	FromChatID int `json:"from_chat_id"`
	MessageID  int `json:"message_id"`
}

// SendPhotoIDQuery ...
type SendPhotoIDQuery struct {
	ChatID           int             `json:"chat_id"`
	Photo            string          `json:"photo"`
	Caption          *string         `json:"caption,omitempty"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendPhotoPathQuery ...
type SendPhotoPathQuery struct {
	ChatID           int             `json:"chat_id"`
	Photo            string          `json:"photo"`
	Caption          *string         `json:"caption,omitempty"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendAudioIDQuery ...
type SendAudioIDQuery struct {
	ChatID           int             `json:"chat_id"`
	Audio            string          `json:"audio"`
	Caption          *string         `json:"caption,omitempty"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendAudioPathQuery ...
type SendAudioPathQuery struct {
	ChatID           int             `json:"chat_id"`
	Audio            string          `json:"audio"`
	Caption          *string         `json:"caption,omitempty"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendDocumentIDQuery ...
type SendDocumentIDQuery struct {
	ChatID           int             `json:"chat_id"`
	Document         string          `json:"document"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendDocumentPathQuery ...
type SendDocumentPathQuery struct {
	ChatID           int             `json:"chat_id"`
	Document         string          `json:"document"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendStickerIDQuery ...
type SendStickerIDQuery struct {
	ChatID           int             `json:"chat_id"`
	Sticker          string          `json:"sticker"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendStickerPathQuery ...
type SendStickerPathQuery struct {
	ChatID           int             `json:"chat_id"`
	Sticker          string          `json:"sticker"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendVideoIDQuery ...
type SendVideoIDQuery struct {
	ChatID           int             `json:"chat_id"`
	Video            string          `json:"video"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendVideoPathQuery ...
type SendVideoPathQuery struct {
	ChatID           int             `json:"chat_id"`
	Video            string          `json:"video"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// SendLocationQuery ...
type SendLocationQuery struct {
	ChatID           int             `json:"chat_id"`
	Latitude         float64         `json:"latitude"`
	Longitude        float64         `json:"longitude"`
	ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
}

// String conversions

func marshall(pay interface{}) string {
	strdata, _ := json.Marshal(pay)
	return string(strdata)
}

func (user User) String() string {
	return marshall(user)
}

func (qsm QuerySendMessage) String() string {
	return marshall(qsm)
}

func (msg Message) String() string {
	return marshall(msg)
}
