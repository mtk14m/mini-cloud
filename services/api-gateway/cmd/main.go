package main

import (
	"log"

	"github.com/mtk14m/mini-cloud/api-gateway/internal/config"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/server"
)

func main() {

	//Charger la configuration
	cfg := config.Load()

	//Créer le serveur
	svr := server.New(cfg)

	//Démarrer le serveur
	log.Printf("API Gateway starting on port %s", cfg.Port)
	if err := svr.Run(); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
