package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func LaunchFirefox(url string) error {
	tryExec := func(path string, args ...string) error {
		cmd := exec.Command(path, args...)
		if err := cmd.Start(); err != nil {
			return err
		}
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		if p, err := exec.LookPath("firefox"); err == nil {
			return tryExec(p, url)
		}
		candidates := []string{
			`C:\\Program Files\\Mozilla Firefox\\firefox.exe`,
			`C:\\Program Files (x86)\\Mozilla Firefox\\firefox.exe`,
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				return tryExec(p, url)
			}
		}

		_ = exec.Command("cmd", "/C", "start", "", url).Start()
		return fmt.Errorf("Firefox introuvable sur ce système Windows")

	case "darwin":
		if err := tryExec("open", "-a", "Firefox", url); err == nil {
			// open -a retourne nil si l'application existe et a été ouverte
			return nil
		}
		_ = exec.Command("open", url).Start()
		return fmt.Errorf("Firefox introuvable via 'open -a Firefox' sur macOS")

	default:
		if p, err := exec.LookPath("firefox"); err == nil {
			return tryExec(p, url)
		}
		if p, err := exec.LookPath("xdg-open"); err == nil {
			_ = exec.Command(p, url).Start()
			return fmt.Errorf("Firefox introuvable sur ce système Linux/Unix")
		}
		return fmt.Errorf("Firefox introuvable et aucun outil de fallback détecté")
	}
}

func main() {
	addr := ":8080"

	// Handler simple qui renvoie une page HTML affichant "Hello World"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>Hello</title></head><body><h1>Hello World</h1></body></html>")
	})

	srvErr := make(chan error, 1)
	go func() {
		log.Printf("Démarrage du serveur sur %s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			srvErr <- err
		}
	}()

	time.Sleep(200 * time.Millisecond)

	fmt.Println("Server running on http://localhost:8080")

	if err := LaunchFirefox("http://localhost:8080"); err != nil {
		fmt.Println("⚠Ouvrez manuellement http://localhost:8080")
		fmt.Printf("Détail: %v\n", err)
	} else {
		fmt.Println("Firefox lancé (ou tentative de lancement réussie)")
	}

	select {
	case err := <-srvErr:
		log.Fatalf("Serveur arrêté avec erreur: %v", err)
	}
}
