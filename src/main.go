package main

import (
	"dle-proxy/database"
	"dle-proxy/database/domain"
	"dle-proxy/server"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("START")

	log.Println("runtime.GOMAXPROCS:", runtime.GOMAXPROCS(0))

	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Cant load .env: ", err)
	}

	mysqlURL := os.Getenv("MYSQL_URL")
	port := os.Getenv("HTTP_PORT")

	if os.Getenv("MYSQL_URL_FILE") != "" {
		mysqlURL_, err := os.ReadFile(os.Getenv("MYSQL_URL_FILE"))
		if err != nil {
			log.Fatal(err)
		}
		mysqlURL = strings.TrimSpace(string(mysqlURL_))
	}

	dbService, err := database.NewService(mysqlURL)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("dbService OK")
	}

	domainService, err := domain.NewService(dbService, 60)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("domainService OK")
	}

	serverService, err := server.NewService(port, dbService, domainService)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("starting server...")
	serverService.Run()

}
