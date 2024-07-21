package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v3"
)

type Room struct {
	ID                 string
	Lock               sync.Mutex
	Title              string
	HostUid            string
	Host               string
	HostName           string
	ClientConnected    bool
	HostConnected      bool
	Offer              webrtc.SessionDescription
	Answer             webrtc.SessionDescription
	HostIceCandidate   []webrtc.ICECandidateInit
	ClientIceCandidate []webrtc.ICECandidateInit
}

var rooms map[string]*Room

var mapLock sync.Mutex

func init() {
	rooms = make(map[string]*Room)
}

type CreateRequset struct {
	Judul    string `json:"judul"`
	Uid      string `json:"uid"`
	Host     string `json:"host"`
	HostName string `json:"host_name"`
}

func CreateRoom() fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := new(CreateRequset)

		if err := c.BodyParser(request); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		roomID := generateRoomID()
		room := &Room{
			ID:       roomID,
			Title:    request.Judul,
			HostUid:  request.Uid,
			Host:     request.Host,
			HostName: request.HostName,
		}
		rooms[roomID] = room
		return c.JSON(room)
	}
}

func GetRooms() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(rooms)
	}
}

func DeleteRoom() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roomID := c.Params("roomID")

		mapLock.Lock()         // Lock the map
		defer mapLock.Unlock() // Ensure the map is unlocked after the operation

		if _, exists := rooms[roomID]; exists {
			delete(rooms, roomID) // Remove the room from the map
		} else {
			fmt.Printf("Room with ID %s does not exist\n", roomID)
		}
		return c.JSON(rooms)
	}
}

func JoinRoom() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roomID := c.Params("roomID")
		uid := c.Params("uid")
		room, ok := rooms[roomID]
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
		}

		// Parse the incoming signal
		var signal struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := c.BodyParser(&signal); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid signal format"})
		}

		// Handle signaling based on signal.Type (offer, answer, candidate, etc.)
		switch signal.Type {
		case "offer":
			// Process offer and create an answer
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal(signal.Data, &offer); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid offer format"})
			}

			room.Lock.Lock()

			room.Offer = offer

			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "offer", "data": room})

		case "answer":
			// Process answer and set remote description
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal(signal.Data, &answer); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid answer format"})
			}

			room.Lock.Lock()

			room.Answer = answer

			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "answer", "data": room})

		case "candidate":
			// Process ICE candidate
			candidate := webrtc.ICECandidateInit{}

			if err := json.Unmarshal(signal.Data, &candidate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ICE candidate format"})
			}

			room.Lock.Lock()
			if room.HostUid == uid {
				room.HostIceCandidate = append(room.HostIceCandidate, candidate)
			} else {
				room.ClientIceCandidate = append(room.ClientIceCandidate, candidate)
			}
			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "candidate", "data": room})
		case "connected":
			// Process ICE candidate
			room.Lock.Lock()
			if room.HostUid == uid {
				room.HostConnected = true
			} else {
				room.ClientConnected = true
			}
			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "candidate", "data": room})
		case "disconnected":

			room.Lock.Lock()
			room.HostIceCandidate = []webrtc.ICECandidateInit{}
			room.ClientIceCandidate = []webrtc.ICECandidateInit{}
			room.Offer = webrtc.SessionDescription{}
			room.Answer = webrtc.SessionDescription{}
			room.HostConnected = false
			room.ClientConnected = false

			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "candidate", "data": room})

		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Unknown signal type"})
		}

		// return c.SendStatus(fiber.StatusOK)
	}
}

func GetOfferAnswer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roomID := c.Params("roomID")
		// uid := c.Params("uid")
		room, ok := rooms[roomID]
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"room": room})
	}
}

func GetCandidate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roomID := c.Params("roomID")
		// uid := c.Params("uid")
		room, ok := rooms[roomID]
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"room": room})
	}
}

func generateRoomID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
