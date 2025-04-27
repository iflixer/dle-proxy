package main

import (
	"dle-proxy/database"
	"dle-proxy/database/domain"
	"dle-proxy/database/domainAlias"
	"dle-proxy/database/domainFile"
	"dle-proxy/database/flixPost"
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

	domainAliasService, err := domainAlias.NewService(dbService, 60)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("domainAliasService OK")
	}

	fileService, err := domainFile.NewService(dbService, 60)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("fileService OK")
	}

	flixPostService, err := flixPost.NewService(dbService, 60)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("flixPost OK")
	}

	serverService, err := server.NewService(port, dbService, domainService, domainAliasService, fileService, flixPostService)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("starting server...")
	serverService.Run()

}
