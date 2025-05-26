package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user account in the social media platform
type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username         string             `bson:"username" json:"username"`
	Email            string             `bson:"email" json:"email"`
	PasswordHash     string             `bson:"password_hash" json:"-"`
	FirstName        string             `bson:"first_name" json:"first_name,omitempty"`
	LastName         string             `bson:"last_name" json:"last_name,omitempty"`
	DisplayName      string             `bson:"display_name" json:"display_name"`
	Bio              string             `bson:"bio" json:"bio,omitempty"`
	ProfilePicture   string             `bson:"profile_picture" json:"profile_picture,omitempty"`
	CoverPhoto       string             `bson:"cover_photo" json:"cover_photo,omitempty"`
	PhoneNumber      string             `bson:"phone_number" json:"phone_number,omitempty"`
	Website          string             `bson:"website" json:"website,omitempty"`
	Location         string             `bson:"location" json:"location,omitempty"`
	DateOfBirth      time.Time          `bson:"date_of_birth" json:"date_of_birth,omitempty"`
	Gender           string             `bson:"gender" json:"gender,omitempty"`
	IsVerified       bool               `bson:"is_verified" json:"is_verified"`
	IsPrivate        bool               `bson:"is_private" json:"is_private"`
	Role             string             `bson:"role" json:"role"`
	Status           string             `bson:"status" json:"status"`
	LastActive       time.Time          `bson:"last_active" json:"last_active"`
	JoinedAt         time.Time          `bson:"joined_at" json:"joined_at"`
	EmailVerified    bool               `bson:"email_verified" json:"email_verified"`
	TwoFactorEnabled bool               `bson:"two_factor_enabled" json:"two_factor_enabled"`
	FollowerCount    int                `bson:"follower_count" json:"follower_count"`
	FollowingCount   int                `bson:"following_count" json:"following_count"`
	PostCount        int                `bson:"post_count" json:"post_count"`
	Settings         UserSettings       `bson:"settings" json:"settings"`
	DeletedAt        *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

// UserSettings contains user customizable settings
type UserSettings struct {
	NotificationPreferences NotificationPreferences `bson:"notification_preferences" json:"notification_preferences"`
	PrivacySettings         PrivacySettings         `bson:"privacy_settings" json:"privacy_settings"`
	LanguagePreference      string                  `bson:"language_preference" json:"language_preference"`
	ThemePreference         string                  `bson:"theme_preference" json:"theme_preference"`
	AutoPlayVideos          bool                    `bson:"auto_play_videos" json:"auto_play_videos"`
	ShowOnlineStatus        bool                    `bson:"show_online_status" json:"show_online_status"`
}

// NotificationPreferences defines what notifications a user receives
type NotificationPreferences struct {
	Push            NotificationTypes `bson:"push" json:"push"`
	Email           NotificationTypes `bson:"email" json:"email"`
	InApp           NotificationTypes `bson:"in_app" json:"in_app"`
	MessagePreviews bool              `bson:"message_previews" json:"message_previews"`
	DoNotDisturb    struct {
		Enabled    bool     `bson:"enabled" json:"enabled"`
		StartTime  string   `bson:"start_time" json:"start_time"`
		EndTime    string   `bson:"end_time" json:"end_time"`
		Exceptions []string `bson:"exceptions" json:"exceptions"`
	} `bson:"do_not_disturb" json:"do_not_disturb"`
}

// NotificationTypes defines which types of notifications are enabled
type NotificationTypes struct {
	Likes          bool `bson:"likes" json:"likes"`
	Comments       bool `bson:"comments" json:"comments"`
	Follows        bool `bson:"follows" json:"follows"`
	Messages       bool `bson:"messages" json:"messages"`
	TaggedPosts    bool `bson:"tagged_posts" json:"tagged_posts"`
	Stories        bool `bson:"stories" json:"stories"`
	LiveStreams    bool `bson:"live_streams" json:"live_streams"`
	GroupActivity  bool `bson:"group_activity" json:"group_activity"`
	EventReminders bool `bson:"event_reminders" json:"event_reminders"`
	SystemUpdates  bool `bson:"system_updates" json:"system_updates"`
}

// PrivacySettings defines the user's privacy preferences
type PrivacySettings struct {
	WhoCanSeeMyPosts        string   `bson:"who_can_see_my_posts" json:"who_can_see_my_posts"`
	WhoCanSendMeMessages    string   `bson:"who_can_send_me_messages" json:"who_can_send_me_messages"`
	WhoCanSeeMyFriends      string   `bson:"who_can_see_my_friends" json:"who_can_see_my_friends"`
	WhoCanTagMe             string   `bson:"who_can_tag_me" json:"who_can_tag_me"`
	WhoCanSeeMyStories      string   `bson:"who_can_see_my_stories" json:"who_can_see_my_stories"`
	BlockedUsers            []string `bson:"blocked_users" json:"blocked_users"`
	MutedUsers              []string `bson:"muted_users" json:"muted_users"`
	HideMyOnlineStatus      bool     `bson:"hide_my_online_status" json:"hide_my_online_status"`
	HideMyLastSeen          bool     `bson:"hide_my_last_seen" json:"hide_my_last_seen"`
	HideMyProfileFromSearch bool     `bson:"hide_my_profile_from_search" json:"hide_my_profile_from_search"`
}
