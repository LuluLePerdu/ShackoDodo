package firefox

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"proxy-interceptor/cert"
	"strings"
)

func makeFirefoxProfile() (string, error) {
	root := filepath.Join(os.TempDir(), "ff-proxy-prof")

	// Supprimer le profil existant pour repartir de zéro
	if err := os.RemoveAll(root); err != nil {
		log.Printf("warning: failed to remove old profile %s: %v", root, err)
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}

	// Force manual HTTP and HTTPS proxy
	userJS := []string{
		// Proxy settings
		`user_pref("network.proxy.type", 1);`,
		`user_pref("network.proxy.http", "127.0.0.1");`,
		`user_pref("network.proxy.http_port", 8181);`,
		`user_pref("network.proxy.ssl", "127.0.0.1");`,
		`user_pref("network.proxy.ssl_port", 8181);`,
		`user_pref("network.proxy.no_proxies_on", "");`,
		`user_pref("network.proxy.bypass_on_local", false);`,
		`user_pref("network.proxy.allow_hijacking_localhost", true);`,

		// CRITICAL: Use Windows system certificates
		`user_pref("security.enterprise_roots.enabled", true);`,

		// Disable certificate pinning completely
		`user_pref("security.cert_pinning.enforcement_level", 0);`,

		// Disable OCSP completely
		`user_pref("security.ssl.enable_ocsp_stapling", false);`,
		`user_pref("security.ssl.enable_ocsp_must_staple", false);`,
		`user_pref("security.OCSP.enabled", 0);`,
		`user_pref("security.OCSP.require", false);`,

		// Allow mixed content
		`user_pref("security.mixed_content.block_active_content", false);`,
		`user_pref("security.mixed_content.block_display_content", false);`,
		`user_pref("security.mixed_content.upgrade_display_content", false);`,

		// Disable MITM detection completely
		`user_pref("security.certerrors.mitm.auto_enable_enterprise_roots", true);`,
		`user_pref("security.certerrors.mitm.priming.enabled", false);`,
		`user_pref("security.ssl.errorReporting.enabled", false);`,
		`user_pref("security.ssl.treat_unsafe_negotiation_as_broken", false);`,
		`user_pref("security.ssl.require_safe_negotiation", false);`,

		// Disable HSTS and cert pinning
		`user_pref("network.stricttransportsecurity.preloadlist", false);`,
		`user_pref("security.cert_pinning.process_headers_from_non_builtin_roots", true);`,

		// Allow deprecated TLS versions
		`user_pref("security.tls.version.enable-deprecated", true);`,
		`user_pref("security.tls.version.min", 1);`,
		`user_pref("security.tls.version.max", 4);`,

		// Disable certificate error telemetry
		`user_pref("security.certerrors.recordEventTelemetry", false);`,
		`user_pref("security.identityblock.show_extended_validation", false);`,

		// Allow weak crypto
		`user_pref("security.ssl3.rsa_des_ede3_sha", true);`,

		// Disable all security warnings
		`user_pref("security.warn_entering_secure", false);`,
		`user_pref("security.warn_leaving_secure", false);`,
		`user_pref("security.warn_submit_insecure", false);`,
		`user_pref("security.insecure_connection_text.enabled", false);`,
		`user_pref("security.insecure_field_warning.contextual.enabled", false);`,

		// Import Windows certificates
		`user_pref("security.osclientcerts.autoload", true);`,

		// CRITICAL: Accept self-signed certificates
		`user_pref("network.http.phishy-userpass-length", 255);`,
		`user_pref("browser.xul.error_pages.expert_bad_cert", true);`,
	}
	if err := os.WriteFile(filepath.Join(root, "user.js"), []byte(strings.Join(userJS, "\n")), 0o644); err != nil {
		return "", err
	}

	// Create a handlers.json file
	handlersJSON := `{
		"defaultHandlersVersion": {"en-US": 4},
		"mimeTypes": {},
		"schemes": {}
	}`
	if err := os.WriteFile(filepath.Join(root, "handlers.json"), []byte(handlersJSON), 0o644); err != nil {
		return "", err
	}

	// Install CA certificate with certutil (NSS) into the profile if available
	if cert.CACertPath != "" && fileExists(cert.CACertPath) {
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)

		// IMPORTANT: Only look for Mozilla NSS certutil, NOT Windows certutil
		// Windows certutil is in system32 and has different options
		potential := []string{
			filepath.Join(exeDir, "certutil.exe"),
			filepath.Join(exeDir, "nss", "bin", "certutil.exe"),
			`C:\Program Files\Mozilla Firefox\certutil.exe`,
			`C:\Program Files\Mozilla Firefox\nss\bin\certutil.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\certutil.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\nss\bin\certutil.exe`,
		}

		// Do NOT use exec.LookPath as it may find Windows certutil.exe

		var certutilPath string
		for _, p := range potential {
			if fileExists(p) {
				certutilPath = p
				break
			}
		}

		if certutilPath != "" {
			log.Printf("Utilisation de certutil: %s", certutilPath)

			// Initialize NSS DB (sql:)
			out, err := exec.Command(certutilPath, "-N", "-d", "sql:"+root, "--empty-password").CombinedOutput()
			if err != nil {
				log.Printf("certutil -N failed (%s): %v\n%s", certutilPath, err, out)
			} else {
				log.Printf("NSS DB initialisée à %s", root)
			}

			// Import trusted CA with proper trust flags
			// CT,C,C = Trust for SSL, Email, and Object Signing
			out2, err2 := exec.Command(certutilPath, "-A", "-n", "ShackoDodo Proxy CA", "-t", "CT,C,C", "-i", cert.CACertPath, "-d", "sql:"+root).CombinedOutput()
			if err2 != nil {
				log.Printf("certutil -A failed (%s): %v\n%s", certutilPath, err2, out2)
			} else {
				log.Printf("Certificat CA importé dans le profil Firefox NSS DB avec trust flags CT,C,C")
			}

			// Verify the certificate was imported
			out3, err3 := exec.Command(certutilPath, "-L", "-d", "sql:"+root).CombinedOutput()
			if err3 == nil {
				log.Printf("Certificats dans le profil Firefox:\n%s", out3)
			}
		} else {
			log.Printf("certutil introuvable; Firefox utilisera security.enterprise_roots.enabled pour faire confiance au magasin Windows")
		}

		// Copy CA certificate to profile for convenience
		if caCertData, err := os.ReadFile(cert.CACertPath); err == nil {
			caCertDest := filepath.Join(root, "shackododo-ca.crt")
			if err := os.WriteFile(caCertDest, caCertData, 0o644); err != nil {
				log.Printf("warning: failed to write CA to profile: %v", err)
			}
		} else {
			log.Printf("warning: failed to read CA cert from %s: %v", cert.CACertPath, err)
		}
	}

	return root, nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Start() {
	prof, err := makeFirefoxProfile()
	if err != nil {
		log.Fatalf("Firefox Profile: %v", err)
	}

	// Obtenir le chemin absolu vers la page de test
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	testPagePath := filepath.Join(exeDir, "..", "test-page.html")

	// Convertir en URL file://
	testPageURL := "file:///" + strings.ReplaceAll(testPagePath, "\\", "/")

	log.Printf("Ouverture de Firefox avec la page de test: %s", testPageURL)

	firefoxPath := `C:\Program Files\Mozilla Firefox\firefox.exe`
	cmd := exec.Command(firefoxPath, "-no-remote", "-profile", prof, testPageURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Cannot start Firefox: %v", err)
	}
}
