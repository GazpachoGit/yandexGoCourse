// set FILE_STORAGE_PATH=../../internal/storage/storage.txt | set SERVER_ADDRESS=:8080 | set BASE_URL=http://localhost:8080/
//shortener -a=:8080 -b=http://localhost:8080/ -f=../../internal/storage/storage.txt
package main

import (
	"fmt"
	"log"
	"net/http"

	serverconfig "github.com/GazpachoGit/yandexGoCourse/internal/config"
	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
)

func main() {

	cfg, err := serverconfig.GetConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("filePath: " + cfg.FilePath)
	fmt.Println("ServerAddres: " + cfg.ServerAddres)
	fmt.Println("BaseURL: " + cfg.BaseURL)

	urlMap, err := storage.NewURLMap(cfg.FilePath)

	defer func() {
		err := urlMap.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if err != nil {
		log.Fatal(err)
		return
	}
	r := handlers.NewShortenerHandler(urlMap, cfg.BaseURL)
	server := &http.Server{
		Addr:    cfg.ServerAddres,
		Handler: r,
	}
	log.Fatal(server.ListenAndServe())

	//http.ListenAndServe(":8080", r)
}
