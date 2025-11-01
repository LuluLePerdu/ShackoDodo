package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

var (
	caCert     *x509.Certificate
	caKey      *rsa.PrivateKey
	CACertPath string
)

// InitCA initializes the Certificate Authority
func InitCA() error {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)

	certPath := filepath.Join(exeDir, "shackododo-ca.crt")
	keyPath := filepath.Join(exeDir, "shackododo-ca.key")
	CACertPath = certPath

	// Check if certificate and key already exist
	if fileExists(certPath) && fileExists(keyPath) {
		// Load existing certificate
		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			return err
		}
		block, _ := pem.Decode(certPEM)
		if block == nil {
			return fmt.Errorf("failed to decode certificate PEM")
		}
		caCert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return err
		}

		// Load existing private key
		keyPEM, err := os.ReadFile(keyPath)
		if err != nil {
			return err
		}
		keyBlock, _ := pem.Decode(keyPEM)
		if keyBlock == nil {
			return fmt.Errorf("failed to decode key PEM")
		}
		caKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return err
		}

		return nil
	}

	// Generate new CA private key
	caKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create CA certificate
	caCert = &x509.Certificate{
		SerialNumber: big.NewInt(2023),
		Subject: pkix.Name{
			Organization:  []string{"ShackoDodo Proxy"},
			Country:       []string{"FR"},
			Province:      []string{""},
			Locality:      []string{"Paris"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "ShackoDodo Proxy CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Self-sign the CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Save CA certificate to project directory
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})

	// Save CA private key to project directory
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GenerateCertForHost generates a certificate for a specific host
func GenerateCertForHost(host string) (*tls.Certificate, error) {
	// Generate private key for the host
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Parse host to check if it's an IP
	var dnsNames []string
	var ipAddresses []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ipAddresses = append(ipAddresses, ip)
	} else {
		dnsNames = append(dnsNames, host)
	}

	// Create certificate template
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"ShackoDodo Proxy"},
			CommonName:   host,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
	}

	// Create certificate signed by CA
	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, &certPrivKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	// Create tls.Certificate
	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  certPrivKey,
	}

	return tlsCert, nil
}
