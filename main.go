package main

import (
	"log"

	"github.com/domyid/chatserver/config"
	"github.com/domyid/chatserver/helper/chatroot"
	"github.com/domyid/chatserver/url"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2"
)

func main() {
	go chatroot.RunHub()

	site := fiber.New(config.Iteung)
	site.Use(cors.New(config.Cors))
	url.Web(site)
	log.Fatal(site.Listen(config.IPPort))
}
