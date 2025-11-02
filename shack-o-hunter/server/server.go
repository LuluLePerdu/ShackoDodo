package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

//go:embed dist
var distFS embed.FS

func Start(port string) {
	// Créer un sous-système de fichiers pour servir depuis /dist
	distSub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Printf("Warning: unable to load embedded frontend: %v", err)
		return
	}

	// Servir les fichiers statiques
	fileServer := http.FileServer(http.FS(distSub))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Si c'est une requête API ou WebSocket, passer
		if r.URL.Path == "/ws" || len(r.URL.Path) > 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}

		// Servir le frontend React
		fileServer.ServeHTTP(w, r)
	})

	go func() {
		log.Printf("Frontend React démarré sur http://localhost:%s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("Erreur serveur HTTP: %v", err)
		}
	}()

	// Attendre un peu que le serveur démarre, puis ouvrir le navigateur
	time.Sleep(500 * time.Millisecond)
	openBrowser("http://localhost:" + port)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		log.Printf("Plateforme non supportée pour l'ouverture automatique du navigateur")
		return
	}
	if err != nil {
		log.Printf("Impossible d'ouvrir le navigateur: %v", err)
	}
}
