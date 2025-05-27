package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Caqil/vyrall/internal/services/livestream"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LiveStreamHandler manages live streaming websocket events
type LiveStreamHandler struct {
	hub               *Hub
	liveStreamService *livestream.Service
}

// NewLiveStreamHandler creates a new live stream handler
func NewLiveStreamHandler(hub *Hub, liveStreamService *livestream.Service) *LiveStreamHandler {
	return &LiveStreamHandler{
		hub:               hub,
		liveStreamService: liveStreamService,
	}
}

// LiveStreamEvent represents a live stream event payload
type LiveStreamEvent struct {
	Type         string `json:"type"`
	StreamID     string `json:"stream_id"`
	Action       string `json:"action,omitempty"`
	Content      string `json:"content,omitempty"`
	ReactionType string `json:"reaction_type,omitempty"`
}

// ProcessStreamEvent handles incoming live stream events
func (h *LiveStreamHandler) ProcessStreamEvent(client *Client, data []byte) {
	var event LiveStreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling live stream event: %v", err)
		client.SendErrorMessage("Invalid event format")
		return
	}

	// Validate stream ID
	if event.StreamID == "" {
		client.SendErrorMessage("Stream ID is required")
		return
	}

	streamID, err := primitive.ObjectIDFromHex(event.StreamID)
	if err != nil {
		client.SendErrorMessage("Invalid stream ID")
		return
	}

	// Process different event types
	switch event.Action {
	case "join":
		h.handleJoinStream(client, streamID)
	case "leave":
		h.handleLeaveStream(client, streamID)
	case "comment":
		h.handleStreamComment(client, streamID, event.Content)
	case "react":
		h.handleStreamReaction(client, streamID, event.ReactionType)
	default:
		client.SendErrorMessage("Unknown stream action")
	}
}

// handleJoinStream processes a user joining a stream
func (h *LiveStreamHandler) handleJoinStream(client *Client, streamID primitive.ObjectID) {
	// Add user to stream viewers
	err := h.liveStreamService.AddStreamViewer(client.ctx, streamID, client.UserID)
	if err != nil {
		log.Printf("Error adding stream viewer: %v", err)
		client.SendErrorMessage("Failed to join stream")
		return
	}

	// Add client to stream room
	roomID := "stream:" + streamID.Hex()
	h.hub.roomsHandler.AddClientToRoom(client, roomID)

	// Notify streamer that a user joined
	stream, err := h.liveStreamService.GetLiveStream(client.ctx, streamID)
	if err != nil {
		log.Printf("Error getting stream: %v", err)
		return
	}

	// Send join notification to stream owner
	joinPayload, err := json.Marshal(map[string]interface{}{
		"type":      "stream_viewer_joined",
		"stream_id": streamID.Hex(),
		"user_id":   client.UserID.Hex(),
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling join payload: %v", err)
		return
	}
	h.hub.SendToUser(stream.UserID, joinPayload)

	// Send stream info to the client
	streamInfo, err := json.Marshal(map[string]interface{}{
		"type":          "stream_joined",
		"stream_id":     streamID.Hex(),
		"stream_status": stream.Status,
		"streamer_id":   stream.UserID.Hex(),
		"viewer_count":  stream.ViewerCount,
		"timestamp":     time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling stream info: %v", err)
		return
	}
	client.SendMessage(streamInfo)
}

// handleLeaveStream processes a user leaving a stream
func (h *LiveStreamHandler) handleLeaveStream(client *Client, streamID primitive.ObjectID) {
	// Remove user from stream viewers
	err := h.liveStreamService.RemoveStreamViewer(client.ctx, streamID, client.UserID)
	if err != nil {
		log.Printf("Error removing stream viewer: %v", err)
		client.SendErrorMessage("Failed to leave stream")
		return
	}

	// Remove client from stream room
	roomID := "stream:" + streamID.Hex()
	h.hub.roomsHandler.RemoveClientFromRoom(client, roomID)

	// Notify streamer that a user left
	stream, err := h.liveStreamService.GetLiveStream(client.ctx, streamID)
	if err != nil {
		log.Printf("Error getting stream: %v", err)
		return
	}

	// Send leave notification to stream owner
	leavePayload, err := json.Marshal(map[string]interface{}{
		"type":      "stream_viewer_left",
		"stream_id": streamID.Hex(),
		"user_id":   client.UserID.Hex(),
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling leave payload: %v", err)
		return
	}
	h.hub.SendToUser(stream.UserID, leavePayload)
}

// handleStreamComment processes a comment on a stream
func (h *LiveStreamHandler) handleStreamComment(client *Client, streamID primitive.ObjectID, content string) {
	if content == "" {
		client.SendErrorMessage("Comment content is required")
		return
	}

	// Save comment in the database
	comment, err := h.liveStreamService.AddStreamComment(client.ctx, streamID, client.UserID, content)
	if err != nil {
		log.Printf("Error adding stream comment: %v", err)
		client.SendErrorMessage("Failed to add comment")
		return
	}

	// Broadcast comment to all stream viewers
	commentPayload, err := json.Marshal(map[string]interface{}{
		"type":       "stream_comment",
		"stream_id":  streamID.Hex(),
		"comment_id": comment.ID.Hex(),
		"user_id":    client.UserID.Hex(),
		"content":    content,
		"timestamp":  comment.CreatedAt,
	})
	if err != nil {
		log.Printf("Error marshaling comment payload: %v", err)
		return
	}

	roomID := "stream:" + streamID.Hex()
	h.hub.roomsHandler.BroadcastToRoom(roomID, commentPayload)
}

// handleStreamReaction processes a reaction to a stream
func (h *LiveStreamHandler) handleStreamReaction(client *Client, streamID primitive.ObjectID, reactionType string) {
	if reactionType == "" {
		client.SendErrorMessage("Reaction type is required")
		return
	}

	// Save reaction in the database
	reaction, err := h.liveStreamService.AddStreamReaction(client.ctx, streamID, client.UserID, reactionType)
	if err != nil {
		log.Printf("Error adding stream reaction: %v", err)
		client.SendErrorMessage("Failed to add reaction")
		return
	}

	// Broadcast reaction to all stream viewers
	reactionPayload, err := json.Marshal(map[string]interface{}{
		"type":          "stream_reaction",
		"stream_id":     streamID.Hex(),
		"reaction_id":   reaction.ID.Hex(),
		"user_id":       client.UserID.Hex(),
		"reaction_type": reactionType,
		"timestamp":     reaction.CreatedAt,
	})
	if err != nil {
		log.Printf("Error marshaling reaction payload: %v", err)
		return
	}

	roomID := "stream:" + streamID.Hex()
	h.hub.roomsHandler.BroadcastToRoom(roomID, reactionPayload)
}

// NotifyStreamStatusChange notifies viewers about a stream status change
func (h *LiveStreamHandler) NotifyStreamStatusChange(streamID primitive.ObjectID, status string) {
	// Get the stream
	stream, err := h.liveStreamService.GetLiveStream(nil, streamID)
	if err != nil {
		log.Printf("Error getting stream: %v", err)
		return
	}

	// Create notification payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":        "stream_status_change",
		"stream_id":   streamID.Hex(),
		"status":      status,
		"streamer_id": stream.UserID.Hex(),
		"timestamp":   time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling status change payload: %v", err)
		return
	}

	// Broadcast to stream room
	roomID := "stream:" + streamID.Hex()
	h.hub.roomsHandler.BroadcastToRoom(roomID, payload)

	// If the stream is starting, also notify followers
	if status == "live" {
		// Notify followers that the stream has started
		h.notifyFollowersOfStreamStart(stream)
	}
}

// notifyFollowersOfStreamStart notifies a user's followers about a stream start
func (h *LiveStreamHandler) notifyFollowersOfStreamStart(stream *models.LiveStream) {
	// Get user's followers
	followers, err := h.liveStreamService.GetUserFollowers(nil, stream.UserID)
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		return
	}

	// Create notification payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":        "streamer_went_live",
		"stream_id":   stream.ID.Hex(),
		"streamer_id": stream.UserID.Hex(),
		"title":       stream.Title,
		"thumbnail":   stream.ThumbnailURL,
		"timestamp":   time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling live notification payload: %v", err)
		return
	}

	// Get follower IDs
	followerIDs := make([]primitive.ObjectID, len(followers))
	for i, follower := range followers {
		followerIDs[i] = follower.FollowerID
	}

	// Broadcast to followers
	h.hub.BroadcastToUsers(followerIDs, payload)
}
