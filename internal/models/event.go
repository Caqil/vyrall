package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents a scheduled event
type Event struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title          string               `bson:"title" json:"title"`
	Description    string               `bson:"description" json:"description"`
	HostID         primitive.ObjectID   `bson:"host_id" json:"host_id"`
	CoHosts        []primitive.ObjectID `bson:"co_hosts,omitempty" json:"co_hosts,omitempty"`
	StartTime      time.Time            `bson:"start_time" json:"start_time"`
	EndTime        time.Time            `bson:"end_time" json:"end_time"`
	TimeZone       string               `bson:"time_zone" json:"time_zone"`
	Location       EventLocation        `bson:"location" json:"location"`
	CoverImage     string               `bson:"cover_image,omitempty" json:"cover_image,omitempty"`
	Type           string               `bson:"type" json:"type"`       // in-person, online, hybrid
	Privacy        string               `bson:"privacy" json:"privacy"` // public, private, invite-only
	GroupID        *primitive.ObjectID  `bson:"group_id,omitempty" json:"group_id,omitempty"`
	Category       string               `bson:"category,omitempty" json:"category,omitempty"`
	Tags           []string             `bson:"tags,omitempty" json:"tags,omitempty"`
	MaxAttendees   int                  `bson:"max_attendees,omitempty" json:"max_attendees,omitempty"`
	RSVPCount      EventRSVPCounts      `bson:"rsvp_count" json:"rsvp_count"`
	IsRecurring    bool                 `bson:"is_recurring" json:"is_recurring"`
	RecurrenceRule string               `bson:"recurrence_rule,omitempty" json:"recurrence_rule,omitempty"`
	ParentEventID  *primitive.ObjectID  `bson:"parent_event_id,omitempty" json:"parent_event_id,omitempty"` // For recurring events
	Status         string               `bson:"status" json:"status"`                                       // scheduled, live, ended, cancelled
	URL            string               `bson:"url,omitempty" json:"url,omitempty"`                         // For online events
	TicketInfo     *EventTicketInfo     `bson:"ticket_info,omitempty" json:"ticket_info,omitempty"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
}

// EventLocation represents an event's location
type EventLocation struct {
	Type        string    `bson:"type" json:"type"` // physical, online, hybrid
	Name        string    `bson:"name,omitempty" json:"name,omitempty"`
	Address     string    `bson:"address,omitempty" json:"address,omitempty"`
	City        string    `bson:"city,omitempty" json:"city,omitempty"`
	State       string    `bson:"state,omitempty" json:"state,omitempty"`
	Country     string    `bson:"country,omitempty" json:"country,omitempty"`
	ZipCode     string    `bson:"zip_code,omitempty" json:"zip_code,omitempty"`
	Coordinates *GeoPoint `bson:"coordinates,omitempty" json:"coordinates,omitempty"`
	OnlineURL   string    `bson:"online_url,omitempty" json:"online_url,omitempty"`
	AccessInfo  string    `bson:"access_info,omitempty" json:"access_info,omitempty"`
}

// EventRSVPCounts tracks the number of different RSVP types
type EventRSVPCounts struct {
	Going      int `bson:"going" json:"going"`
	Interested int `bson:"interested" json:"interested"`
	NotGoing   int `bson:"not_going" json:"not_going"`
	NoReply    int `bson:"no_reply" json:"no_reply"`
	Waitlist   int `bson:"waitlist" json:"waitlist"`
}

// EventAttendee represents a user's RSVP to an event
type EventAttendee struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	EventID          primitive.ObjectID  `bson:"event_id" json:"event_id"`
	UserID           primitive.ObjectID  `bson:"user_id" json:"user_id"`
	RSVP             string              `bson:"rsvp" json:"rsvp"` // going, interested, not_going
	RSVPTimestamp    time.Time           `bson:"rsvp_timestamp" json:"rsvp_timestamp"`
	CheckedIn        bool                `bson:"checked_in" json:"checked_in"`
	CheckInTime      *time.Time          `bson:"check_in_time,omitempty" json:"check_in_time,omitempty"`
	InvitedBy        *primitive.ObjectID `bson:"invited_by,omitempty" json:"invited_by,omitempty"`
	GuestCount       int                 `bson:"guest_count" json:"guest_count"`
	Note             string              `bson:"note,omitempty" json:"note,omitempty"`
	TicketType       string              `bson:"ticket_type,omitempty" json:"ticket_type,omitempty"`
	TicketID         string              `bson:"ticket_id,omitempty" json:"ticket_id,omitempty"`
	IsWaitlisted     bool                `bson:"is_waitlisted" json:"is_waitlisted"`
	WaitlistPosition int                 `bson:"waitlist_position,omitempty" json:"waitlist_position,omitempty"`
}

// EventTicketInfo contains information about event tickets
type EventTicketInfo struct {
	IsPaid        bool              `bson:"is_paid" json:"is_paid"`
	Currency      string            `bson:"currency,omitempty" json:"currency,omitempty"`
	TicketTypes   []EventTicketType `bson:"ticket_types,omitempty" json:"ticket_types,omitempty"`
	SalesStartAt  time.Time         `bson:"sales_start_at" json:"sales_start_at"`
	SalesEndAt    time.Time         `bson:"sales_end_at" json:"sales_end_at"`
	RefundPolicy  string            `bson:"refund_policy,omitempty" json:"refund_policy,omitempty"`
	TaxPercentage float64           `bson:"tax_percentage,omitempty" json:"tax_percentage,omitempty"`
	ServiceFee    float64           `bson:"service_fee,omitempty" json:"service_fee,omitempty"`
}

// EventTicketType defines a type of ticket for an event
type EventTicketType struct {
	ID          string    `bson:"id" json:"id"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
	Price       float64   `bson:"price" json:"price"`
	Quantity    int       `bson:"quantity" json:"quantity"`
	Available   int       `bson:"available" json:"available"`
	Sold        int       `bson:"sold" json:"sold"`
	SalesEndAt  time.Time `bson:"sales_end_at,omitempty" json:"sales_end_at,omitempty"`
	IsHidden    bool      `bson:"is_hidden" json:"is_hidden"`
}

// EventReminder represents a reminder for an event
type EventReminder struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventID      primitive.ObjectID `bson:"event_id" json:"event_id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ReminderTime time.Time          `bson:"reminder_time" json:"reminder_time"`
	Type         string             `bson:"type" json:"type"`     // email, push, sms
	Status       string             `bson:"status" json:"status"` // pending, sent, failed
	SentAt       *time.Time         `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}
