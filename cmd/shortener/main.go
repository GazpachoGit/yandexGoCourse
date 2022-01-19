// set FILE_STORAGE_PATH=../../internal/storage/storage.txt | set SERVER_ADDRESS=:8080 | set BASE_URL=http://localhost:8080/

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	serverConfig "github.com/GazpachoGit/yandexGoCourse/internal/config"
	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
)

func main() {

	str := os.Getenv("FILE_STORAGE_PATH")
	fmt.Println("filePath")
	fmt.Println(str)

	cfg, err := serverConfig.GetConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("server")
	fmt.Println(cfg.ServerAddres)

	urlMap, err := storage.NewUrlMap(cfg.FilePath)
	defer urlMap.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	r := handlers.NewShortenerHandler(urlMap, cfg.BaseUrl)
	server := &http.Server{
		Addr:    cfg.ServerAddres,
		Handler: r,
	}
	log.Fatal(server.ListenAndServe())

	//http.ListenAndServe(":8080", r)
}
