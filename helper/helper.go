package helper

import (
	"log"
	"os"
	"strings"
)

func GetAddress() (ipport string, network string) {
	port := os.Getenv("PORT")
	log.Println("SERVER PORT >> ", port)
	network = "tcp4"
	if port == "" {
		ipport = ":5199"
	} else if port[0:1] != ":" {
		ip := os.Getenv("IP")
		log.Println("SERVER IP >> ", ip)
		if ip == "" {
			ipport = ":" + port
		} else {
			if strings.Contains(ip, ".") {
				ipport = ip + ":" + port
			} else {
				ipport = "[" + ip + "]" + ":" + port
				network = "tcp6"
			}
		}
	}

	return
}
