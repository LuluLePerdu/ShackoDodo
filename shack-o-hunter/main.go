package main

import (
	"fmt"
	"log"
	"proxy-interceptor/admin"
	"proxy-interceptor/browsers"
	"proxy-interceptor/cert"
	"proxy-interceptor/config"
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

	// Détecter et lancer les navigateurs disponibles
	available := browsers.DetectAvailableBrowsers()

	if len(available) == 0 {
		log.Println("Aucun navigateur supporté trouvé")
		log.Println("Navigateurs supportés: Firefox, Chrome, Edge")
	} else {
		log.Printf("Navigateurs détectés: %v", available)

		if len(available) == 1 {
			// Un seul navigateur disponible, le lancer directement
			browser := available[0]
			log.Printf("Lancement de %s...", browser.String())
			if err := browsers.StartBrowser(browser); err != nil {
				log.Printf("Erreur lors du lancement de %s: %v", browser.String(), err)
			}
		} else {
			// Plusieurs navigateurs disponibles, afficher les options
			fmt.Println("\nPlusieurs navigateurs détectés:")
			for i, browser := range available {
				fmt.Printf("  %d. %s\n", i+1, browser.String())
			}
			fmt.Printf("  %d. Lancer tous les navigateurs\n", len(available)+1)
			fmt.Print("\nChoisissez une option (1-" + fmt.Sprintf("%d", len(available)+1) + ") ou appuyez sur Entrée pour le premier: ")

			var choice int
			fmt.Scanln(&choice)

			if choice == 0 || choice == 1 {
				// Par défaut ou premier choix
				browser := available[0]
				log.Printf("Lancement de %s...", browser.String())
				if err := browsers.StartBrowser(browser); err != nil {
					log.Printf("Erreur lors du lancement de %s: %v", browser.String(), err)
				}
			} else if choice > 1 && choice <= len(available) {
				// Navigateur spécifique choisi
				browser := available[choice-1]
				log.Printf("Lancement de %s...", browser.String())
				if err := browsers.StartBrowser(browser); err != nil {
					log.Printf("Erreur lors du lancement de %s: %v", browser.String(), err)
				}
			} else if choice == len(available)+1 {
				// Lancer tous les navigateurs
				log.Println("Lancement de tous les navigateurs disponibles...")
				errors := browsers.StartAllAvailableBrowsers()
				if len(errors) > 0 {
					for _, err := range errors {
						log.Printf("Erreur: %v", err)
					}
				}
			} else {
				log.Printf("Choix invalide, lancement de %s par défaut...", available[0].String())
				if err := browsers.StartBrowser(available[0]); err != nil {
					log.Printf("Erreur lors du lancement de %s: %v", available[0].String(), err)
				}
			}
		}
	}

	fmt.Println("\nProxy ShackoDodo démarré!")
	fmt.Println("- Ouvrez payload-modifier.html dans un navigateur pour l'interface de modification")
	fmt.Println("- Appuyez sur Ctrl+C pour arrêter")

	select {}
}
