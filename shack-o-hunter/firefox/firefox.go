package firefox

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func makeFirefoxProfile() (string, error) {
	root := filepath.Join(os.TempDir(), "ff-proxy-prof")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	// Force manual HTTP proxy only
	// network.proxy.type = 1 (manual)
	// network.proxy.http = 127.0.0.1
	// network.proxy.http_port = 8181
	// network.proxy.no_proxies_on = "" (proxy everything)
	// Disable proxy for HTTPS to keep this test strictly HTTP-side
	userJS := []string{
		`user_pref("network.proxy.type", 1);`,
		`user_pref("network.proxy.http", "127.0.0.1");`,
		`user_pref("network.proxy.http_port", 8181);`,
		`user_pref("network.proxy.no_proxies_on", "");`,
		`user_pref("network.proxy.bypass_on_local", false);`,
		`user_pref("network.proxy.allow_hijacking_localhost", true);`,
		// keep HTTPS unset if you only handle plain HTTP:
		`user_pref("network.proxy.ssl", "");`,
		`user_pref("network.proxy.ssl_port", 0);`,
	}
	if err := os.WriteFile(filepath.Join(root, "user.js"), []byte(strings.Join(userJS, "\n")), 0o644); err != nil {
		return "", err
	}
	return root, nil
}

func Start() {
	// 2) create a temp firefox profile with proxy forced
	prof, err := makeFirefoxProfile()
	if err != nil {
		log.Fatalf("Firefox Profile: %v", err)
	}

	// 3) launch firefox with that profile (no env var needed)
	firefoxPath := `C:\Program Files\Mozilla Firefox\firefox.exe`
	cmd := exec.Command(firefoxPath, "-no-remote", "-profile", prof, "http://example.com")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Cannot start Firefox: %v", err)
	}
}
