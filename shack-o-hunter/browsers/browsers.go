package browsers

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"proxy-interceptor/cert"
	"strings"
)

type Browser int

const (
	Firefox Browser = iota
	Chrome
	Edge
)

func (b Browser) String() string {
	switch b {
	case Firefox:
		return "Firefox"
	case Chrome:
		return "Chrome"
	case Edge:
		return "Edge"
	default:
		return "Unknown"
	}
}

type BrowserPaths struct {
	ExecutablePaths []string
	ProfileDir      string
	CertUtilPaths   []string
}

var browserConfigs = map[Browser]BrowserPaths{
	Firefox: {
		ExecutablePaths: []string{
			`C:\Program Files\Mozilla Firefox\firefox.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\firefox.exe`,
		},
		ProfileDir: "ff-proxy-prof",
		CertUtilPaths: []string{
			`C:\Program Files\Mozilla Firefox\certutil.exe`,
			`C:\Program Files\Mozilla Firefox\nss\bin\certutil.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\certutil.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\nss\bin\certutil.exe`,
		},
	},
	Chrome: {
		ExecutablePaths: []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Users\` + os.Getenv("USERNAME") + `\AppData\Local\Google\Chrome\Application\chrome.exe`,
		},
		ProfileDir: "chrome-proxy-prof",
	},
	Edge: {
		ExecutablePaths: []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		},
		ProfileDir: "edge-proxy-prof",
	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findBrowserExecutable(browser Browser) string {
	config := browserConfigs[browser]
	for _, path := range config.ExecutablePaths {
		if fileExists(path) {
			return path
		}
	}
	return ""
}

func DetectAvailableBrowsers() []Browser {
	var available []Browser
	for browser := range browserConfigs {
		if findBrowserExecutable(browser) != "" {
			available = append(available, browser)
		}
	}
	return available
}

func createFirefoxProfile(profilePath string) error {
	userJS := []string{
		`user_pref("network.proxy.type", 1);`,
		`user_pref("network.proxy.http", "127.0.0.1");`,
		`user_pref("network.proxy.http_port", 8181);`,
		`user_pref("network.proxy.ssl", "127.0.0.1");`,
		`user_pref("network.proxy.ssl_port", 8181);`,
		`user_pref("network.proxy.no_proxies_on", "");`,
		`user_pref("network.proxy.bypass_on_local", false);`,
		`user_pref("network.proxy.allow_hijacking_localhost", true);`,
		`user_pref("security.enterprise_roots.enabled", true);`,
		`user_pref("security.cert_pinning.enforcement_level", 0);`,
		`user_pref("security.ssl.enable_ocsp_stapling", false);`,
		`user_pref("security.ssl.enable_ocsp_must_staple", false);`,
		`user_pref("security.OCSP.enabled", 0);`,
		`user_pref("security.OCSP.require", false);`,
		`user_pref("security.mixed_content.block_active_content", false);`,
		`user_pref("security.mixed_content.block_display_content", false);`,
		`user_pref("security.mixed_content.upgrade_display_content", false);`,
		`user_pref("security.certerrors.mitm.auto_enable_enterprise_roots", true);`,
		`user_pref("security.certerrors.mitm.priming.enabled", false);`,
		`user_pref("security.ssl.errorReporting.enabled", false);`,
		`user_pref("security.ssl.treat_unsafe_negotiation_as_broken", false);`,
		`user_pref("security.ssl.require_safe_negotiation", false);`,
		`user_pref("network.stricttransportsecurity.preloadlist", false);`,
		`user_pref("security.cert_pinning.process_headers_from_non_builtin_roots", true);`,
		`user_pref("security.tls.version.enable-deprecated", true);`,
		`user_pref("security.tls.version.min", 1);`,
		`user_pref("security.tls.version.max", 4);`,
		`user_pref("security.certerrors.recordEventTelemetry", false);`,
		`user_pref("security.identityblock.show_extended_validation", false);`,
		`user_pref("security.ssl3.rsa_des_ede3_sha", true);`,
		`user_pref("security.warn_entering_secure", false);`,
		`user_pref("security.warn_leaving_secure", false);`,
		`user_pref("security.warn_submit_insecure", false);`,
		`user_pref("security.insecure_connection_text.enabled", false);`,
		`user_pref("security.insecure_field_warning.contextual.enabled", false);`,
		`user_pref("security.osclientcerts.autoload", true);`,
		`user_pref("network.http.phishy-userpass-length", 255);`,
		`user_pref("browser.xul.error_pages.expert_bad_cert", true);`,
	}

	if err := os.WriteFile(filepath.Join(profilePath, "user.js"), []byte(strings.Join(userJS, "\n")), 0o644); err != nil {
		return err
	}

	handlersJSON := `{
		"defaultHandlersVersion": {"en-US": 4},
		"mimeTypes": {},
		"schemes": {}
	}`
	return os.WriteFile(filepath.Join(profilePath, "handlers.json"), []byte(handlersJSON), 0o644)
}

func setupChromeProfile(profilePath string) error {
	prefsJSON := `{
		"proxy": {
			"mode": "fixed_servers",
			"server": "127.0.0.1:8181"
		},
		"ssl": {
			"version_min": "tls1",
			"version_max": "tls1.3"
		}
	}`

	defaultDir := filepath.Join(profilePath, "Default")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(defaultDir, "Preferences"), []byte(prefsJSON), 0o644)
}

func createBrowserProfile(browser Browser) (string, error) {
	config := browserConfigs[browser]
	profilePath := filepath.Join(os.TempDir(), config.ProfileDir)

	if err := os.RemoveAll(profilePath); err != nil {
		log.Printf("Warning: failed to remove old profile %s: %v", profilePath, err)
	}

	if err := os.MkdirAll(profilePath, 0o755); err != nil {
		return "", err
	}

	switch browser {
	case Firefox:
		if err := createFirefoxProfile(profilePath); err != nil {
			return "", err
		}
		installFirefoxCertificate(profilePath)
	case Chrome, Edge:
		if err := setupChromeProfile(profilePath); err != nil {
			return "", err
		}
	}

	return profilePath, nil
}

func installFirefoxCertificate(profilePath string) {
	if cert.CACertPath == "" || !fileExists(cert.CACertPath) {
		return
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	certutilPaths := []string{
		filepath.Join(exeDir, "certutil.exe"),
		filepath.Join(exeDir, "nss", "bin", "certutil.exe"),
	}
	certutilPaths = append(certutilPaths, browserConfigs[Firefox].CertUtilPaths...)

	var certutilPath string
	for _, p := range certutilPaths {
		if fileExists(p) {
			certutilPath = p
			break
		}
	}

	if certutilPath != "" {
		log.Printf("Using certutil: %s", certutilPath)

		exec.Command(certutilPath, "-N", "-d", "sql:"+profilePath, "--empty-password").Run()
		exec.Command(certutilPath, "-A", "-n", "ShackoDodo Proxy CA", "-t", "CT,C,C", "-i", cert.CACertPath, "-d", "sql:"+profilePath).Run()

		log.Printf("Certificate imported into Firefox NSS DB")
	} else {
		log.Printf("certutil not found; Firefox will use security.enterprise_roots.enabled")
	}

	if caCertData, err := os.ReadFile(cert.CACertPath); err == nil {
		os.WriteFile(filepath.Join(profilePath, "shackododo-ca.crt"), caCertData, 0o644)
	}
}

func StartBrowser(browser Browser) error {
	executable := findBrowserExecutable(browser)
	if executable == "" {
		return fmt.Errorf("%s not found", browser.String())
	}

	profilePath, err := createBrowserProfile(browser)
	if err != nil {
		return fmt.Errorf("failed to create %s profile: %v", browser.String(), err)
	}

	startURL := "http://www.example.com"

	var cmd *exec.Cmd
	switch browser {
	case Firefox:
		cmd = exec.Command(executable, "-no-remote", "-profile", profilePath, startURL)
	case Chrome:
		cmd = exec.Command(executable,
			"--user-data-dir="+profilePath,
			"--proxy-server=127.0.0.1:8181",
			"--ignore-certificate-errors",
			"--ignore-ssl-errors",
			"--ignore-certificate-errors-spki-list",
			"--disable-web-security",
			"--allow-running-insecure-content",
			"--disable-features=VizDisplayCompositor",
			startURL)
	case Edge:
		cmd = exec.Command(executable,
			"--user-data-dir="+profilePath,
			"--proxy-server=127.0.0.1:8181",
			"--ignore-certificate-errors",
			"--ignore-ssl-errors",
			"--ignore-certificate-errors-spki-list",
			"--disable-web-security",
			"--allow-running-insecure-content",
			startURL)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func StartDefaultBrowser() error {
	available := DetectAvailableBrowsers()
	if len(available) == 0 {
		return fmt.Errorf("no supported browsers found")
	}

	browser := available[0]
	log.Printf("Detected browsers: %v, starting %s", available, browser.String())
	return StartBrowser(browser)
}

func StartAllAvailableBrowsers() []error {
	available := DetectAvailableBrowsers()
	if len(available) == 0 {
		return []error{fmt.Errorf("no supported browsers found")}
	}

	var errors []error
	for _, browser := range available {
		if err := StartBrowser(browser); err != nil {
			errors = append(errors, fmt.Errorf("%s: %v", browser.String(), err))
		}
	}

	return errors
}
