package url

import (
	"github.com/domyid/chatserver/handler"
	"github.com/domyid/chatserver/helper/chatroot"
	"github.com/domyid/chatserver/helper/wrtc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func Web(page *fiber.App) {

	page.Get("/ws", websocket.New(chatroot.RunSocket))
	page.Get("/webrtc", websocket.New(wrtc.RunWebRTCSocket)) // New route for WebRTC signaling

	// route for new webrtc server
	// Endpoint untuk membuat room baru
	page.Post("/api/create-room", handler.CreateRoom())
	page.Get("/api/rooms", handler.GetRooms())
	// Endpoint untuk meng-handle WebRTC signaling
	page.Post("/room/:roomID/signal", handler.JoinRoom())

}
