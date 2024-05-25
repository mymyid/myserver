package url

import (
	"github.com/domyid/chatserver/helper/chatroot"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func Web(page *fiber.App) {

	page.Get("/ws", websocket.New(chatroot.RunSocket))
}
