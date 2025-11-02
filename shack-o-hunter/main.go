package main

import (
	"log"
	"proxy-interceptor/admin"
	"proxy-interceptor/cert"
	"proxy-interceptor/config"
	"proxy-interceptor/firefox"
	"proxy-interceptor/proxy"
	"proxy-interceptor/websocket"
	"time"
)

func main() {
	// Vérifier si on est admin, sinon demander l'élévation
	if !admin.IsAdmin() {
		log.Println("Le programme nécessite des privilèges administrateur pour installer le certificat CA.")
		log.Println("Demande d'élévation...")
		if err := admin.RequestElevation(); err != nil {
			log.Fatalf("Impossible d'obtenir les privilèges administrateur: %v", err)
		}
		return
	}

	log.Println("Démarrage avec privilèges administrateur")

	// Initialiser le certificat CA AVANT de démarrer le proxy
	if err := cert.InitCA(); err != nil {
		log.Fatalf("Erreur initialisation CA: %v", err)
	}
	log.Println("Certificat CA généré/chargé avec succès")

	// Installer le certificat CA dans le magasin système Windows AVANT tout
	if cert.CACertPath != "" {
		if !admin.IsCertInstalledInSystemStore("ShackoDodo Proxy CA") {
			log.Println("Installation du certificat CA dans le magasin système Windows...")
			if err := admin.InstallCertToSystemStore(cert.CACertPath); err != nil {
				log.Printf("Avertissement: impossible d'installer le certificat dans le magasin système: %v", err)
			} else {
				log.Println("Certificat CA installé dans le magasin système Windows (Trusted Root)")
			}
		} else {
			log.Println("Certificat CA déjà présent dans le magasin système")
		}
	}

	// Petit délai pour s'assurer que Windows a bien traité le certificat
	time.Sleep(1 * time.Second)

	// Get config instance
	cfg := config.GetInstance()

	// Maintenant démarrer le proxy
	proxy.Start()
	log.Printf("Proxy démarré sur 127.0.0.1:%d", cfg.ProxyPort)

	// Démarrer le WebSocket
	websocket.Start()
	log.Println("WebSocket démarré")

	// Délai pour s'assurer que tout est prêt
	time.Sleep(1 * time.Second)

	// Enfin, démarrer Firefox avec le profil configuré
	log.Println("Lancement de Firefox...")
	firefox.Start()

	select {}
}
