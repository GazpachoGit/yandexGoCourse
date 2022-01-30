// set FILE_STORAGE_PATH=../../internal/storage/storage.txt | set SERVER_ADDRESS=:8080 | set BASE_URL=http://localhost:8080/
//shortener -a=:8080 -b=http://localhost:8080/ -f=../../internal/storage/storage.txt
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

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

	log.Println("result filePath: " + cfg.FilePath)
	log.Println("result ServerAddres: " + cfg.ServerAddres)
	log.Println("result BaseURL: " + cfg.BaseURL)

	db, err := storage.InitDb(cfg.DBConnectionString)
	defer db.Close()
	if err != nil {
		log.Fatalln(err)
		return
	}
	r, err := handlers.NewShortenerHandler(db, cfg.BaseURL)
	if err != nil {
		log.Fatalln(err)
		return
	}
	server := &http.Server{
		Addr:    cfg.ServerAddres,
		Handler: r,
	}
	ch := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)

		<-sigint
		log.Println("interrupt is catched")

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		log.Println("server stopped")
		close(ch)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
	<-ch
}
