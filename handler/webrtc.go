package handler

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v3"
)

type CommunicationData struct {
	Uid       string
	Offer     webrtc.SessionDescription
	Answer    webrtc.SessionDescription
	Candidate webrtc.ICECandidateInit
}

type Room struct {
	ID          string
	WebRTCConns map[string]*webrtc.PeerConnection // Simpan koneksi WebRTC untuk setiap pengguna di room
	Lock        sync.Mutex
	Title       string
	HostUid     string
	HostData    CommunicationData
	ClientData  CommunicationData
}

var rooms map[string]*Room

func init() {
	rooms = make(map[string]*Room)
}

type CreateRequset struct {
	Judul string `json:"judul"`
	Uid   string `json:"uid"`
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
			HostUid:     request.Uid,
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

		var peerConnection *webrtc.PeerConnection
		var err error

		// Create a new PeerConnection if it doesn't exist
		room.Lock.Lock()
		if _, exists := room.WebRTCConns[c.IP()]; !exists {
			peerConnection, err = createPeerConnection()
			if err != nil {
				room.Lock.Unlock()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create PeerConnection"})
			}
			room.WebRTCConns[c.IP()] = peerConnection
		} else {
			peerConnection = room.WebRTCConns[c.IP()]
		}
		room.Lock.Unlock()

		// Handle signaling based on signal.Type (offer, answer, candidate, etc.)
		switch signal.Type {
		case "offer":
			// Process offer and create an answer
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal(signal.Data, &offer); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid offer format"})
			}

			room.Lock.Lock()

			if room.HostUid == uid {
				room.HostData = CommunicationData{Uid: uid, Offer: offer, Answer: room.HostData.Answer, Candidate: room.HostData.Candidate}
			} else {
				room.ClientData = CommunicationData{Uid: uid, Offer: offer, Answer: room.ClientData.Answer, Candidate: room.ClientData.Candidate}
			}

			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "offer", "data": offer})

			// if err := peerConnection.SetRemoteDescription(offer); err != nil {
			// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set remote description"})
			// }

			// answer, err := peerConnection.CreateAnswer(nil)
			// if err != nil {
			// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create answer"})
			// }

			// if err := peerConnection.SetLocalDescription(answer); err != nil {
			// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set local description"})
			// }

			// return c.JSON(fiber.Map{"type": "answer", "data": answer})
		case "answer":
			// Process answer and set remote description
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal(signal.Data, &answer); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid answer format"})
			}

			room.Lock.Lock()
			if room.HostUid == uid {
				room.HostData = CommunicationData{Uid: uid, Answer: answer, Offer: room.HostData.Offer, Candidate: room.HostData.Candidate}
			} else {
				room.ClientData = CommunicationData{Uid: uid, Answer: answer, Offer: room.ClientData.Offer, Candidate: room.ClientData.Candidate}
			}
			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "answer", "data": answer})

			// if err := peerConnection.SetRemoteDescription(answer); err != nil {
			// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set remote description"})
			// }
		case "candidate":
			// Process ICE candidate
			candidate := webrtc.ICECandidateInit{}

			if err := json.Unmarshal(signal.Data, &candidate); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ICE candidate format"})
			}

			room.Lock.Lock()
			if room.HostUid == uid {
				room.HostData = CommunicationData{Uid: uid, Answer: room.HostData.Answer, Offer: room.HostData.Offer, Candidate: candidate}
			} else {
				room.ClientData = CommunicationData{Uid: uid, Answer: room.ClientData.Answer, Offer: room.ClientData.Offer, Candidate: candidate}
			}
			room.Lock.Unlock()
			return c.JSON(fiber.Map{"type": "candidate", "data": candidate})

			// if err := peerConnection.AddICECandidate(candidate); err != nil {
			// 	log.Println("ERR >> ", err.Error())
			// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add ICE candidate"})
			// }
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

// Helper function to create a new PeerConnection
func createPeerConnection() (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	return webrtc.NewPeerConnection(config)
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
