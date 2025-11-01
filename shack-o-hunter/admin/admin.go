package admin

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32       = syscall.NewLazyDLL("shell32.dll")
	shellExecuteW = shell32.NewProc("ShellExecuteW")
)

// IsAdmin checks if the current process is running with administrator privileges
func IsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

// RequestElevation restarts the program with administrator privileges
func RequestElevation() error {
	verb := "runas"
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1

	ret, _, _ := shellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		uintptr(showCmd),
	)

	if ret <= 32 {
		return syscall.Errno(ret)
	}

	os.Exit(0)
	return nil
}

// InstallCertToSystemStore installs the CA certificate to Windows Trusted Root store
func InstallCertToSystemStore(certPath string) error {
	log.Printf("Installation du certificat CA dans le magasin système Windows...")

	cmd := exec.Command("certutil", "-addstore", "-f", "Root", certPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Erreur lors de l'installation du certificat: %v\nSortie: %s", err, output)
		return err
	}

	log.Printf("Certificat CA installé avec succès dans le magasin système")
	log.Printf("Sortie certutil: %s", output)
	return nil
}

// UninstallCertFromSystemStore removes the CA certificate from Windows Trusted Root store
func UninstallCertFromSystemStore(certName string) error {
	log.Printf("Suppression du certificat CA du magasin système...")

	cmd := exec.Command("certutil", "-delstore", "Root", certName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Avertissement lors de la suppression du certificat: %v\nSortie: %s", err, output)
		return err
	}

	log.Printf("Certificat CA supprimé du magasin système")
	return nil
}

// IsCertInstalledInSystemStore checks if the certificate is already installed
func IsCertInstalledInSystemStore(certName string) bool {
	cmd := exec.Command("certutil", "-verifystore", "Root", certName)
	err := cmd.Run()
	return err == nil
}
