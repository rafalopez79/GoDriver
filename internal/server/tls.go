package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

//NewServerTLSConfig creates tls config for server
func NewServerTLSConfig(caPem, certPem, keyPem []byte, authType tls.ClientAuthType) (config *tls.Config, err error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPem) {
		return nil, fmt.Errorf("Error appending CA")
	}
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, err
	}
	config = &tls.Config{
		ClientAuth:   authType,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
	}
	return config, nil
}

// extract RSA public key from certificate
func getPublicKeyFromCert(certPem []byte) (pk []byte, err error) {
	block, _ := pem.Decode(certPem)
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey, err := x509.MarshalPKIXPublicKey(crt.PublicKey.(*rsa.PublicKey))
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKey}), nil
}

// generate and sign RSA certificates with given CA
func generateAndSignRSACerts(caPem, caKey []byte) (certPem []byte, keyPem []byte, err error) {
	// Load CA
	catls, err := tls.X509KeyPair(caPem, caKey)
	if err != nil {
		return nil, nil, err
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		return nil, nil, err
	}
	// use the CA to sign certificates
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"ORGANIZATION_NAME"},
			Country:       []string{"COUNTRY_CODE"},
			Province:      []string{"PROVINCE"},
			Locality:      []string{"CITY"},
			StreetAddress: []string{"ADDRESS"},
			PostalCode:    []string{"POSTAL_CODE"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	// sign the certificate
	certb, err := x509.CreateCertificate(rand.Reader, ca, cert, &priv.PublicKey, catls.PrivateKey)
	if err != nil {
		return nil, nil, err
	}
	certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certb})
	keyPem = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return certPem, keyPem, nil
}

func generateCA() (caPem []byte, caKey []byte, err error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"Organization"},
			Country:       []string{"Country"},
			Province:      []string{"Province"},
			Locality:      []string{"Locality"},
			StreetAddress: []string{"StreetAddress"},
			PostalCode:    []string{"PostalCode"},
		},
		NotBefore:             time.Now().AddDate(0, 0, -1),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	caPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	caKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return caPem, caKey, nil
}
