package main

import (
	"log"

	"github.com/domyid/chatserver/config"
	"github.com/domyid/chatserver/helper/chatroot"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	site := fiber.New(config.Iteung)
	site.Use(cors.New(config.Cors))

	//app := fiber.New()

	site.Static("/", "./home.html")

	go chatroot.RunHub()

	site.Get("/ws", websocket.New(chatroot.RunSocket))

	log.Fatal(site.Listen(config.IPPort))
}
