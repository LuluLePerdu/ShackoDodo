package main

import (
	"fmt"
	"log"
	"proxy-interceptor/admin"
	"proxy-interceptor/browsers"
	"proxy-interceptor/cert"
	"proxy-interceptor/config"
	"proxy-interceptor/proxy"
	"proxy-interceptor/server"
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

	// Démarrer le serveur frontend React
	server.Start("3000")

	// Délai pour s'assurer que tout est prêt
	time.Sleep(1 * time.Second)

	// Démarrer le gestionnaire de lancement de navigateurs
	go handleBrowserLaunches()

	// Afficher les navigateurs disponibles
	available := browsers.DetectAvailableBrowsers()
	if len(available) > 0 {
		log.Printf("Navigateurs détectés: %v", available)
	} else {
		log.Println("Aucun navigateur supporté trouvé")
	}

	fmt.Println("\nProxy ShackoDodo démarré!")
	fmt.Println("- Interface web: http://localhost:3000")
	fmt.Println("- Le navigateur va s'ouvrir automatiquement")
	fmt.Println("- Navigateurs supportés: Firefox, Chrome, Edge")
	fmt.Println("- Appuyez sur Ctrl+C pour arrêter")

	select {}
}

// Gestionnaire pour les demandes de lancement de navigateur depuis l'UI
func handleBrowserLaunches() {
	for request := range websocket.BrowserLaunchChannel {
		log.Printf("Launching browser: %s", request.Browser)

		var browserType browsers.Browser
		switch request.Browser {
		case "firefox":
			browserType = browsers.Firefox
		case "chrome":
			browserType = browsers.Chrome
		case "edge":
			browserType = browsers.Edge
		case "all":
			// Lancer tous les navigateurs disponibles
			available := browsers.DetectAvailableBrowsers()
			for _, browser := range available {
				go func(b browsers.Browser) {
					if err := browsers.StartBrowser(b); err != nil {
						log.Printf("Error launching %s: %v", b.String(), err)
					} else {
						log.Printf("Successfully launched %s", b.String())
					}
				}(browser)
			}
			continue
		default:
			log.Printf("Unknown browser: %s", request.Browser)
			continue
		}

		// Lancer le navigateur spécifique
		if err := browsers.StartBrowser(browserType); err != nil {
			log.Printf("Error launching %s: %v", browserType.String(), err)
		} else {
			log.Printf("Successfully launched %s", browserType.String())
		}
	}
}
