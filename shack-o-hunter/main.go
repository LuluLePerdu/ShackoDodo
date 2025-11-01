package main

import (
	"log"
	"proxy-interceptor/admin"
	"proxy-interceptor/cert"
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

	// Mode MITM complet pour intercepter HTTPS
	proxy.Start()

	// Petit délai pour s'assurer que le certificat CA est bien généré
	time.Sleep(500 * time.Millisecond)

	// Installer le certificat CA dans le magasin système Windows
	if cert.CACertPath != "" {
		if !admin.IsCertInstalledInSystemStore("ShackoDodo Proxy CA") {
			if err := admin.InstallCertToSystemStore(cert.CACertPath); err != nil {
				log.Printf("Avertissement: impossible d'installer le certificat dans le magasin système: %v", err)
			} else {
				log.Println("Certificat CA installé dans le magasin système Windows (Trusted Root)")
			}
		} else {
			log.Println("Certificat CA déjà présent dans le magasin système")
		}
	}

	websocket.Start()
	firefox.Start()

	select {}
}
