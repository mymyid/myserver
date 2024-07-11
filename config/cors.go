package config

import (
	"strings"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

var origins = []string{
	"https://naskah.bukupedia.co.id",
	"https://auth.ulbi.ac.id",
	"https://sip.ulbi.ac.id",
	"https://euis.ulbi.ac.id",
	"https://home.ulbi.ac.id",
	"https://alpha.ulbi.ac.id",
	"https://dias.ulbi.ac.id",
	"https://meet.my.my.id",
	"https://chatgpl.do.my.id",
	"https://www.do.my.id",
	"https://roo.my.my.id",
}

var headers = []string{
	"Origin",
	"Content-Type",
	"Accept",
	"Authorization",
	"Access-Control-Request-Headers",
	"Token",
	"Login",
	"Access-Control-Allow-Origin",
	"Bearer",
	"X-Requested-With",
}

var Cors = cors.Config{
	AllowOrigins:     strings.Join(origins[:], ","),
	AllowHeaders:     strings.Join(headers[:], ","),
	ExposeHeaders:    "Content-Length",
	AllowCredentials: true,
}
