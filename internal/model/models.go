package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base model with UUID primary key
type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"default:now()"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Enum types
type UserStatus string

const (
	UserStatusOnline    UserStatus = "online"
	UserStatusOffline   UserStatus = "offline"
	UserStatusAway      UserStatus = "away"
	UserStatusBusy      UserStatus = "busy"
	UserStatusInvisible UserStatus = "invisible"
)

type ContactStatus string

const (
	ContactStatusPending  ContactStatus = "pending"
	ContactStatusAccepted ContactStatus = "accepted"
	ContactStatusBlocked  ContactStatus = "blocked"
	ContactStatusRejected ContactStatus = "rejected"
)

// UserProfile model for additional user information
type UserProfile struct {
	BaseModel
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index;uniqueIndex"`
	FirstName   string     `json:"first_name" gorm:"size:100"`
	LastName    string     `json:"last_name" gorm:"size:100"`
	DisplayName string     `json:"display_name" gorm:"size:200"`
	Bio         string     `json:"bio" gorm:"type:text"`
	Location    string     `json:"location" gorm:"size:255"`
	Website     string     `json:"website" gorm:"size:500"`
	Company     string     `json:"company" gorm:"size:255"`
	JobTitle    string     `json:"job_title" gorm:"size:255"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string     `json:"gender" gorm:"size:20"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserContact model for managing user contacts/friends
type UserContact struct {
	BaseModel
	UserID    uuid.UUID     `json:"user_id" gorm:"type:uuid;not null;index"`
	ContactID uuid.UUID     `json:"contact_id" gorm:"type:uuid;not null;index"`
	Status    ContactStatus `json:"status" gorm:"size:20;not null;default:'pending'"`
	NickName  string        `json:"nickname" gorm:"size:255"`
	Notes     string        `json:"notes" gorm:"type:text"`

	// Relationships
	User    User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Contact User `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
}

// User model with complete fields based on ERD
type User struct {
	BaseModel
	Username    string     `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Email       string     `json:"email" gorm:"size:255;uniqueIndex;not null"`
	Password    string     `json:"-" gorm:"not null"`
	FirstName   string     `json:"first_name" gorm:"size:100"`
	LastName    string     `json:"last_name" gorm:"size:100"`
	Avatar      string     `json:"avatar" gorm:"size:500"`
	PhoneNumber string     `json:"phone_number" gorm:"size:20"`
	Bio         string     `json:"bio" gorm:"type:text"`
	Status      string     `json:"status" gorm:"size:20;default:'offline'"` // online, offline, away, busy, invisible
	LastSeen    *time.Time `json:"last_seen"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	IsVerified  bool       `json:"is_verified" gorm:"default:false"`

	// User Settings (embedded)
	Language            string `json:"language" gorm:"size:10;default:'en'"`
	Timezone            string `json:"timezone" gorm:"size:50;default:'UTC'"`
	NotificationSound   bool   `json:"notification_sound" gorm:"default:true"`
	EmailNotifications  bool   `json:"email_notifications" gorm:"default:true"`
	PushNotifications   bool   `json:"push_notifications" gorm:"default:true"`
	ShowOnlineStatus    bool   `json:"show_online_status" gorm:"default:true"`
	ShowReadReceipts    bool   `json:"show_read_receipts" gorm:"default:true"`
	AllowDirectMessages bool   `json:"allow_direct_messages" gorm:"default:true"`
	AutoJoinPublicRooms bool   `json:"auto_join_public_rooms" gorm:"default:false"`

	// Relationships
	Sessions      []UserSession  `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	RoomMembers   []RoomMember   `json:"room_members,omitempty" gorm:"foreignKey:UserID"`
	Messages      []Message      `json:"messages,omitempty" gorm:"foreignKey:SenderID"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"foreignKey:UserID"`
}

// UserSession model for managing user sessions and tokens
type UserSession struct {
	BaseModel
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	DeviceID     string    `json:"device_id" gorm:"size:255;not null;index"`
	DeviceType   string    `json:"device_type" gorm:"size:50"` // web, mobile, desktop
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"size:500"`
	AccessToken  string    `json:"access_token" gorm:"size:500;not null;index"`
	RefreshToken string    `json:"refresh_token" gorm:"size:500;not null"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null;index"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Room model for chat rooms/channels
type Room struct {
	BaseModel
	Name        string `json:"name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	Type        string `json:"type" gorm:"size:20;not null;index"` // direct, group, public, broadcast
	Avatar      string `json:"avatar" gorm:"size:500"`
	IsPublic    bool   `json:"is_public" gorm:"default:true;index"`
	MaxMembers  int    `json:"max_members"`

	// Room Settings (embedded)
	AllowFileUpload      bool `json:"allow_file_upload" gorm:"default:true"`
	AllowVoiceMessages   bool `json:"allow_voice_messages" gorm:"default:true"`
	AllowVideoMessages   bool `json:"allow_video_messages" gorm:"default:true"`
	MessageRetentionDays int  `json:"message_retention_days" gorm:"default:0"` // 0 = forever
	RequireApproval      bool `json:"require_approval" gorm:"default:false"`
	MuteAllMembers       bool `json:"mute_all_members" gorm:"default:false"`
	OnlyAdminCanPost     bool `json:"only_admin_can_post" gorm:"default:false"`

	CreatedBy uuid.UUID `json:"created_by" gorm:"type:uuid;not null;index"`

	// Relationships
	CreatedByUser User         `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	Members       []RoomMember `json:"members,omitempty" gorm:"foreignKey:RoomID"`
	Messages      []Message    `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
	Invites       []RoomInvite `json:"invites,omitempty" gorm:"foreignKey:RoomID"`
}

// RoomMember model for room membership
type RoomMember struct {
	BaseModel
	RoomID     uuid.UUID  `json:"room_id" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Role       string     `json:"role" gorm:"size:20;default:'member'"` // owner, admin, moderator, member
	JoinedAt   time.Time  `json:"joined_at" gorm:"default:now()"`
	LastReadAt *time.Time `json:"last_read_at"`
	IsMuted    bool       `json:"is_muted" gorm:"default:false"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	InvitedBy  *uuid.UUID `json:"invited_by" gorm:"type:uuid;index"` // Who invited this user

	// Relationships
	Room          Room  `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	User          User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	InvitedByUser *User `json:"invited_by_user,omitempty" gorm:"foreignKey:InvitedBy"`
}

// Message model for chat messages
type Message struct {
	BaseModel
	RoomID    uuid.UUID  `json:"room_id" gorm:"type:uuid;not null;index"`
	SenderID  uuid.UUID  `json:"sender_id" gorm:"type:uuid;not null;index"`
	ReplyToID *uuid.UUID `json:"reply_to_id" gorm:"type:uuid;index"`
	Type      string     `json:"type" gorm:"size:20;not null;index"` // text, image, video, audio, file, location, system, sticker, voice_note, video_call, audio_call
	Content   string     `json:"content" gorm:"type:text"`
	Metadata  string     `json:"metadata" gorm:"type:jsonb"` // mentioned_users, hashtags, links, location, call_data, system_event
	IsEdited  bool       `json:"is_edited" gorm:"default:false"`
	EditedAt  *time.Time `json:"edited_at"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false"`

	// Relationships
	Room        Room                `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Sender      User                `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	ReplyTo     *Message            `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
	Attachments []MessageAttachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`
	Reactions   []MessageReaction   `json:"reactions,omitempty" gorm:"foreignKey:MessageID"`
	Reads       []MessageRead       `json:"reads,omitempty" gorm:"foreignKey:MessageID"`
}

// MessageAttachment model for file attachments
type MessageAttachment struct {
	BaseModel
	MessageID    uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	FileName     string    `json:"file_name" gorm:"size:255;not null"`
	FileSize     int64     `json:"file_size" gorm:"not null"`
	FileType     string    `json:"file_type" gorm:"size:100;not null;index"`
	MimeType     string    `json:"mime_type" gorm:"size:100;not null;index"`
	URL          string    `json:"url" gorm:"size:500;not null"`
	ThumbnailURL string    `json:"thumbnail_url" gorm:"size:500"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     int       `json:"duration"` // for audio/video in seconds

	// Relationships
	Message Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
}

// MessageReaction model for emoji reactions
type MessageReaction struct {
	BaseModel
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Emoji     string    `json:"emoji" gorm:"size:50;not null;index"`

	// Relationships
	Message Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// MessageRead model for read receipts
type MessageRead struct {
	BaseModel
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ReadAt    time.Time `json:"read_at" gorm:"default:now();index"`

	// Relationships
	Message Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Notification model for user notifications
type Notification struct {
	BaseModel
	UserID  uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Type    string     `json:"type" gorm:"size:50;not null;index"` // message, mention, room_invite, room_join, room_leave, system, call, friend
	Title   string     `json:"title" gorm:"size:255;not null"`
	Message string     `json:"message" gorm:"type:text;not null"`
	Data    string     `json:"data" gorm:"type:jsonb"` // notification specific data
	IsRead  bool       `json:"is_read" gorm:"default:false;index"`
	ReadAt  *time.Time `json:"read_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserBlock model for blocking users
type UserBlock struct {
	BaseModel
	BlockerID uuid.UUID `json:"blocker_id" gorm:"type:uuid;not null;index"`
	BlockedID uuid.UUID `json:"blocked_id" gorm:"type:uuid;not null;index"`
	Reason    string    `json:"reason" gorm:"type:text"`

	// Relationships
	Blocker User `json:"blocker,omitempty" gorm:"foreignKey:BlockerID"`
	Blocked User `json:"blocked,omitempty" gorm:"foreignKey:BlockedID"`
}

// ActivityLog model for user activity tracking
type ActivityLog struct {
	BaseModel
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	ActivityType string     `json:"activity_type" gorm:"size:50;not null;index"` // login, logout, message_sent, room_join, room_leave, room_create, file_upload, user_block, user_unblock
	Description  string     `json:"description" gorm:"type:text"`
	Metadata     string     `json:"metadata" gorm:"type:jsonb"` // activity specific data
	IPAddress    string     `json:"ip_address" gorm:"size:45"`
	UserAgent    string     `json:"user_agent" gorm:"size:500"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// RoomInvite model for room invitations
type RoomInvite struct {
	BaseModel
	RoomID      uuid.UUID  `json:"room_id" gorm:"type:uuid;not null;index"`
	InviterID   uuid.UUID  `json:"inviter_id" gorm:"type:uuid;not null;index"`
	InviteeID   *uuid.UUID `json:"invitee_id" gorm:"type:uuid;index"`             // Optional - for direct invites
	InviteCode  string     `json:"invite_code" gorm:"size:50;unique;index"`       // For shareable links
	Status      string     `json:"status" gorm:"size:20;default:'pending';index"` // pending, accepted, rejected, expired
	Message     string     `json:"message" gorm:"type:text"`
	ExpiresAt   *time.Time `json:"expires_at" gorm:"index"`
	MaxUses     int        `json:"max_uses" gorm:"default:0"` // 0 = unlimited
	UsedCount   int        `json:"used_count" gorm:"default:0"`
	RespondedAt *time.Time `json:"responded_at"`

	// Relationships
	Room    Room  `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Inviter User  `json:"inviter,omitempty" gorm:"foreignKey:InviterID"`
	Invitee *User `json:"invitee,omitempty" gorm:"foreignKey:InviteeID"`
}

// MessageDraft model for message drafts
type MessageDraft struct {
	BaseModel
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	RoomID    uuid.UUID  `json:"room_id" gorm:"type:uuid;not null;index"`
	Content   string     `json:"content" gorm:"type:text;not null"`
	ReplyToID *uuid.UUID `json:"reply_to_id" gorm:"type:uuid"`

	// Relationships
	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Room    Room     `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	ReplyTo *Message `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
}

// TypingIndicator model for typing indicators
type TypingIndicator struct {
	BaseModel
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	RoomID    uuid.UUID `json:"room_id" gorm:"type:uuid;not null;index"`
	StartedAt time.Time `json:"started_at" gorm:"default:now()"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Room Room `json:"room,omitempty" gorm:"foreignKey:RoomID"`
}

// FileUpload model for file uploads
type FileUpload struct {
	BaseModel
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	OriginalName string     `json:"original_name" gorm:"size:255;not null"`
	FileName     string     `json:"file_name" gorm:"size:255;not null;index"`
	FilePath     string     `json:"file_path" gorm:"size:500;not null"`
	FileSize     int64      `json:"file_size" gorm:"not null"`
	FileType     string     `json:"file_type" gorm:"size:100;not null;index"`
	MimeType     string     `json:"mime_type" gorm:"size:100;not null"`
	UploadStatus string     `json:"upload_status" gorm:"size:20;default:'uploading';index"` // uploading, completed, failed, deleted
	IsTemporary  bool       `json:"is_temporary" gorm:"default:true;index"`
	ExpiresAt    *time.Time `json:"expires_at" gorm:"index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserPreferences model for user preferences
type UserPreferences struct {
	BaseModel
	UserID               uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	Theme                string    `json:"theme" gorm:"size:20;default:'light'"`      // light, dark, auto
	FontSize             string    `json:"font_size" gorm:"size:10;default:'medium'"` // small, medium, large
	ChatWallpaper        string    `json:"chat_wallpaper" gorm:"size:500"`
	AutoDownloadMedia    bool      `json:"auto_download_media" gorm:"default:true"`
	CompressImages       bool      `json:"compress_images" gorm:"default:true"`
	KeyboardShortcuts    bool      `json:"keyboard_shortcuts" gorm:"default:true"`
	ShowTypingIndicators bool      `json:"show_typing_indicators" gorm:"default:true"`
	GroupNotifications   bool      `json:"group_notifications" gorm:"default:true"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ServerStats model for server statistics
type ServerStats struct {
	BaseModel
	ServerID           string    `json:"server_id" gorm:"size:100;not null;index"`
	ActiveConnections  int       `json:"active_connections" gorm:"default:0"`
	TotalMessagesToday int       `json:"total_messages_today" gorm:"default:0"`
	TotalUsersOnline   int       `json:"total_users_online" gorm:"default:0"`
	MemoryUsage        int64     `json:"memory_usage" gorm:"default:0"`
	CPUUsage           float64   `json:"cpu_usage" gorm:"type:decimal(5,2);default:0"`
	LastUpdated        time.Time `json:"last_updated" gorm:"default:now();index"`
}

// Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaginatedResponse struct {
	APIResponse
	Meta PaginationMeta `json:"meta"`
}

// Request structures for User Management
type CreateUserRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Bio         string `json:"bio,omitempty"`
}

type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	DeviceID   string `json:"device_id" validate:"required"`
	DeviceType string `json:"device_type,omitempty"` // web, mobile, desktop
}

type UpdateUserRequest struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Bio         string `json:"bio,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Status      string `json:"status,omitempty"`
}

type UpdateUserSettingsRequest struct {
	Language            string `json:"language,omitempty"`
	Timezone            string `json:"timezone,omitempty"`
	NotificationSound   *bool  `json:"notification_sound,omitempty"`
	EmailNotifications  *bool  `json:"email_notifications,omitempty"`
	PushNotifications   *bool  `json:"push_notifications,omitempty"`
	ShowOnlineStatus    *bool  `json:"show_online_status,omitempty"`
	ShowReadReceipts    *bool  `json:"show_read_receipts,omitempty"`
	AllowDirectMessages *bool  `json:"allow_direct_messages,omitempty"`
	AutoJoinPublicRooms *bool  `json:"auto_join_public_rooms,omitempty"`
}

// Request structures for Room Management
type CreateRoomRequest struct {
	Name            string `json:"name" validate:"required,max=255"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type" validate:"required,oneof=direct group public broadcast"`
	Avatar          string `json:"avatar,omitempty"`
	IsPublic        *bool  `json:"is_public,omitempty"`
	MaxMembers      int    `json:"max_members,omitempty"`
	RequireApproval bool   `json:"require_approval,omitempty"`
}

type UpdateRoomRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	IsPublic    *bool  `json:"is_public,omitempty"`
	MaxMembers  int    `json:"max_members,omitempty"`
}

type CreateInviteRequest struct {
	ExpiresIn int `json:"expires_in,omitempty"` // seconds
	MaxUses   int `json:"max_uses,omitempty"`   // 0 = unlimited
}

type JoinRoomRequest struct {
	RoomID uuid.UUID `json:"room_id" validate:"required"`
}

type InviteToRoomRequest struct {
	RoomID    uuid.UUID  `json:"room_id" validate:"required"`
	UserID    uuid.UUID  `json:"user_id" validate:"required"`
	Message   string     `json:"message,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Request structures for Messaging
type SendMessageRequest struct {
	RoomID    uuid.UUID  `json:"room_id" validate:"required"`
	Content   string     `json:"content" validate:"required"`
	Type      string     `json:"type,omitempty" validate:"oneof=text image video audio file location system sticker voice_note"`
	ReplyToID *uuid.UUID `json:"reply_to_id,omitempty"`
	Metadata  string     `json:"metadata,omitempty"`
}

type EditMessageRequest struct {
	Content  string `json:"content" validate:"required"`
	Metadata string `json:"metadata,omitempty"`
}

type ReactToMessageRequest struct {
	Emoji string `json:"emoji" validate:"required"`
}

type MarkAsReadRequest struct {
	MessageID uuid.UUID `json:"message_id" validate:"required"`
}

// Request structures for File Upload
type FileUploadRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required"`
	FileType    string `json:"file_type" validate:"required"`
	MimeType    string `json:"mime_type" validate:"required"`
	IsTemporary bool   `json:"is_temporary,omitempty"`
}

// WebSocket Message Types
type WSMessageType string

const (
	WSTypePing             WSMessageType = "ping"
	WSTypePong             WSMessageType = "pong"
	WSTypeAuth             WSMessageType = "auth"
	WSTypeMessage          WSMessageType = "message"
	WSTypeMessageEdit      WSMessageType = "message_edit"
	WSTypeMessageDelete    WSMessageType = "message_delete"
	WSTypeMessageReaction  WSMessageType = "message_reaction"
	WSTypeTypingStart      WSMessageType = "typing_start"
	WSTypeTypingStop       WSMessageType = "typing_stop"
	WSTypeUserJoin         WSMessageType = "user_join"
	WSTypeUserLeave        WSMessageType = "user_leave"
	WSTypeUserStatusChange WSMessageType = "user_status_change"
	WSTypeRoomJoin         WSMessageType = "room_join"
	WSTypeRoomLeave        WSMessageType = "room_leave"
	WSTypeNotification     WSMessageType = "notification"
	WSTypeError            WSMessageType = "error"
)

// WebSocket Message Structure
type WSMessage struct {
	Type      WSMessageType `json:"type"`
	Data      interface{}   `json:"data,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	ID        string        `json:"id,omitempty"`
}

// WebSocket Authentication
type WSAuthRequest struct {
	Token string `json:"token" validate:"required"`
}

// WebSocket Typing Indicator
type WSTypingRequest struct {
	RoomID uuid.UUID `json:"room_id" validate:"required"`
}

// WebSocket User Status
type WSUserStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=online offline away busy invisible"`
}

// Response structures for Authentication
type LoginResponse struct {
	User         User      `json:"user"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Response structures for Rooms
type RoomWithMembersResponse struct {
	Room
	MemberCount  int        `json:"member_count"`
	UnreadCount  int        `json:"unread_count"`
	LastMessage  *Message   `json:"last_message,omitempty"`
	LastActivity *time.Time `json:"last_activity,omitempty"`
}

type RoomMemberResponse struct {
	RoomMember
	UnreadCount int `json:"unread_count"`
}

// Response structures for Messages
type MessageResponse struct {
	Message
	SenderName    string         `json:"sender_name"`
	SenderAvatar  string         `json:"sender_avatar"`
	ReactionCount map[string]int `json:"reaction_count,omitempty"`
	IsRead        bool           `json:"is_read"`
}

// Notification Response
type NotificationResponse struct {
	Notification
	RelatedUser *User `json:"related_user,omitempty"`
	RelatedRoom *Room `json:"related_room,omitempty"`
}
