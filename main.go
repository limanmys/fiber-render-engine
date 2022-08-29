package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/limanmys/render-engine/pkg/utils"
)

func main() {
	err := godotenv.Load("../server/.env")
	if err != nil {
		log.Fatalln("Cannot read Liman environment file")
	}

	utils.CreateServer(2806)
}
