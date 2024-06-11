package handler

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v3"
)

type Room struct {
	ID          string
	WebRTCConns map[string]*webrtc.PeerConnection // Simpan koneksi WebRTC untuk setiap pengguna di room
	Lock        sync.Mutex
	Title       string
}

var rooms map[string]*Room

func init() {
	rooms = make(map[string]*Room)
}

type CreateRequset struct {
	Judul string `json:"judul"`
}

func CreateRoom() fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := new(CreateRequset)

		if err := c.BodyParser(request); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		roomID := generateRoomID()
		room := &Room{
			ID:          roomID,
			WebRTCConns: make(map[string]*webrtc.PeerConnection),
			Title:       request.Judul,
		}
		rooms[roomID] = room
		return c.JSON(fiber.Map{"roomID": roomID, "title": request.Judul})
	}
}

func GetRooms() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(rooms)
	}
}

func JoinRoom() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roomID := c.Params("roomID")
		_, ok := rooms[roomID]
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
		}

		// TODO: fix this section
		var signal struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := c.BodyParser(&signal); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid signal format"})
		}

		// Handle signaling based on signal.Type (offer, answer, candidate, etc.)
		switch signal.Type {
		case "offer", "answer":
			// Process offer/answer and set remote description
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal(signal.Data, &offer); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid offer format"})
			}
			// Example: room.HandleOfferOrAnswer(signal.Data)
		case "candidate":
			// Process ICE candidate
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal(signal.Data, &candidate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ICE candidate format"})
			}
			// Example: room.HandleICECandidate(signal.Data)
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Unknown signal type"})
		}

		return c.SendStatus(fiber.StatusOK)
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
