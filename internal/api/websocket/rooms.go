package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoomsHandler manages room-based websocket communications
type RoomsHandler struct {
	hub        *Hub
	rooms      map[string]map[*Client]bool
	roomsMutex sync.RWMutex
}

// NewRoomsHandler creates a new rooms handler
func NewRoomsHandler(hub *Hub) *RoomsHandler {
	return &RoomsHandler{
		hub:   hub,
		rooms: make(map[string]map[*Client]bool),
	}
}

// JoinRoom handles a client joining a room
func (h *RoomsHandler) JoinRoom(client *Client, data []byte) {
	var payload struct {
		RoomID string `json:"room_id"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Error unmarshaling join room: %v", err)
		client.SendErrorMessage("Invalid join room format")
		return
	}

	if payload.RoomID == "" {
		client.SendErrorMessage("Room ID is required")
		return
	}

	// Add client to room
	h.AddClientToRoom(client, payload.RoomID)

	// Send acknowledgment to client
	ackPayload, err := json.Marshal(map[string]interface{}{
		"type":    "room_joined",
		"room_id": payload.RoomID,
	})
	if err != nil {
		log.Printf("Error marshaling room joined ack: %v", err)
		return
	}

	client.SendMessage(ackPayload)
}

// LeaveRoom handles a client leaving a room
func (h *RoomsHandler) LeaveRoom(client *Client, data []byte) {
	var payload struct {
		RoomID string `json:"room_id"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Error unmarshaling leave room: %v", err)
		client.SendErrorMessage("Invalid leave room format")
		return
	}

	if payload.RoomID == "" {
		client.SendErrorMessage("Room ID is required")
		return
	}

	// Remove client from room
	h.RemoveClientFromRoom(client, payload.RoomID)

	// Send acknowledgment to client
	ackPayload, err := json.Marshal(map[string]interface{}{
		"type":    "room_left",
		"room_id": payload.RoomID,
	})
	if err != nil {
		log.Printf("Error marshaling room left ack: %v", err)
		return
	}

	client.SendMessage(ackPayload)
}

// AddClientToRoom adds a client to a room
func (h *RoomsHandler) AddClientToRoom(client *Client, roomID string) {
	h.roomsMutex.Lock()
	defer h.roomsMutex.Unlock()

	// Create room if it doesn't exist
	if _, exists := h.rooms[roomID]; !exists {
		h.rooms[roomID] = make(map[*Client]bool)
	}

	// Add client to room
	h.rooms[roomID][client] = true

	log.Printf("Client %s joined room %s", client.UserID.Hex(), roomID)
}

// RemoveClientFromRoom removes a client from a room
func (h *RoomsHandler) RemoveClientFromRoom(client *Client, roomID string) {
	h.roomsMutex.Lock()
	defer h.roomsMutex.Unlock()

	// Check if room exists
	room, exists := h.rooms[roomID]
	if !exists {
		return
	}

	// Remove client from room
	delete(room, client)

	// Remove room if empty
	if len(room) == 0 {
		delete(h.rooms, roomID)
	}

	log.Printf("Client %s left room %s", client.UserID.Hex(), roomID)
}

// BroadcastToRoom broadcasts a message to all clients in a room
func (h *RoomsHandler) BroadcastToRoom(roomID string, message []byte) {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	// Check if room exists
	room, exists := h.rooms[roomID]
	if !exists {
		return
	}

	// Send message to all clients in the room
	for client := range room {
		select {
		case client.Send <- message:
		default:
			// If the client's send buffer is full, assume it's dead and remove from room
			go func(c *Client, rid string) {
				h.RemoveClientFromRoom(c, rid)
				h.hub.Unregister <- c
			}(client, roomID)
		}
	}
}

// GetClientsInRoom gets all clients in a room
func (h *RoomsHandler) GetClientsInRoom(roomID string) []*Client {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	// Check if room exists
	room, exists := h.rooms[roomID]
	if !exists {
		return nil
	}

	// Get all clients in the room
	clients := make([]*Client, 0, len(room))
	for client := range room {
		clients = append(clients, client)
	}

	return clients
}

// GetUserIDsInRoom gets all user IDs in a room
func (h *RoomsHandler) GetUserIDsInRoom(roomID string) []primitive.ObjectID {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	// Check if room exists
	room, exists := h.rooms[roomID]
	if !exists {
		return nil
	}

	// Get all user IDs in the room
	userIDs := make([]primitive.ObjectID, 0, len(room))
	for client := range room {
		userIDs = append(userIDs, client.UserID)
	}

	return userIDs
}

// RemoveClientFromAllRooms removes a client from all rooms
func (h *RoomsHandler) RemoveClientFromAllRooms(client *Client) {
	h.roomsMutex.Lock()
	defer h.roomsMutex.Unlock()

	// Remove client from all rooms
	for roomID, room := range h.rooms {
		if _, exists := room[client]; exists {
			delete(room, client)

			// Remove room if empty
			if len(room) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}
}

// IsClientInRoom checks if a client is in a room
func (h *RoomsHandler) IsClientInRoom(client *Client, roomID string) bool {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	// Check if room exists
	room, exists := h.rooms[roomID]
	if !exists {
		return false
	}

	// Check if client is in room
	_, clientExists := room[client]
	return clientExists
}
