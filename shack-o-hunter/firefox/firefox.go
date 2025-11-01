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
	os.RemoveAll(root)

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

		// Disable ALL certificate validation (critical for MITM)
		`user_pref("security.enterprise_roots.enabled", true);`,
		`user_pref("security.cert_pinning.enforcement_level", 0);`,
		`user_pref("security.ssl.enable_ocsp_stapling", false);`,
		`user_pref("security.ssl.enable_ocsp_must_staple", false);`,
		`user_pref("security.OCSP.enabled", 0);`,
		`user_pref("security.OCSP.require", false);`,
		`user_pref("security.mixed_content.block_active_content", false);`,
		`user_pref("security.mixed_content.block_display_content", false);`,

		// Disable MITM detection completely
		`user_pref("security.certerrors.mitm.auto_enable_enterprise_roots", true);`,
		`user_pref("security.certerrors.mitm.priming.enabled", false);`,
		`user_pref("security.ssl.errorReporting.enabled", false);`,
		`user_pref("security.ssl.treat_unsafe_negotiation_as_broken", false);`,
		`user_pref("security.ssl.require_safe_negotiation", false);`,

		// Accept ALL invalid certificates
		`user_pref("network.stricttransportsecurity.preloadlist", false);`,
		`user_pref("security.cert_pinning.process_headers_from_non_builtin_roots", false);`,
		`user_pref("security.tls.version.enable-deprecated", true);`,

		// Disable certificate verification errors
		`user_pref("security.certerrors.recordEventTelemetry", false);`,
		`user_pref("security.identityblock.show_extended_validation", false);`,

		// Allow weak crypto (for testing)
		`user_pref("security.ssl3.rsa_des_ede3_sha", true);`,
		`user_pref("security.tls.version.min", 1);`,
		`user_pref("security.tls.version.max", 4);`,

		// Disable security warnings
		`user_pref("security.warn_entering_secure", false);`,
		`user_pref("security.warn_leaving_secure", false);`,
		`user_pref("security.warn_submit_insecure", false);`,
		`user_pref("security.insecure_connection_text.enabled", false);`,
		`user_pref("security.insecure_field_warning.contextual.enabled", false);`,
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
	os.WriteFile(filepath.Join(root, "handlers.json"), []byte(handlersJSON), 0o644)

	// Install CA certificate with certutil if available
	if cert.CACertPath != "" && fileExists(cert.CACertPath) {
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)

		certutilPaths := []string{
			filepath.Join(exeDir, "certutil.exe"),
			`C:\Program Files\Mozilla Firefox\certutil.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\certutil.exe`,
		}

		var certutilPath string
		for _, path := range certutilPaths {
			if fileExists(path) {
				certutilPath = path
				break
			}
		}

		if certutilPath != "" {
			cmdInit := exec.Command(certutilPath, "-N", "-d", "sql:"+root, "--empty-password")
			cmdInit.Run()

			cmdImport := exec.Command(certutilPath, "-A", "-n", "ShackoDodo Proxy CA",
				"-t", "C,,", "-i", cert.CACertPath, "-d", "sql:"+root)
			cmdImport.Run()
		}

		// Copy CA certificate to profile
		caCertData, err := os.ReadFile(cert.CACertPath)
		if err == nil {
			caCertDest := filepath.Join(root, "shackododo-ca.crt")
			os.WriteFile(caCertDest, caCertData, 0o644)
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

	firefoxPath := `C:\Program Files\Mozilla Firefox\firefox.exe`
	cmd := exec.Command(firefoxPath, "-no-remote", "-profile", prof, "http://example.com")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Cannot start Firefox: %v", err)
	}
}
