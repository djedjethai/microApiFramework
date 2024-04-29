package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type tlsConfig struct {
	certFile      string
	keyFile       string
	caFile        string
	serverAddress string
	server        bool
}

func configFileTLS(pathToTLS, filename string) string {
	if pathToTLS != "" {
		return filepath.Join(pathToTLS, filename)
	}
	return filepath.Join(pathToTLSDefault, filename)
}

func setupTLSConfig(cfg tlsConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}
	if cfg.certFile != "" && cfg.keyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(
			cfg.certFile,
			cfg.keyFile,
		)
		if err != nil {
			return nil, err
		}
	}
	if cfg.caFile != "" {
		b, err := ioutil.ReadFile(cfg.caFile)
		if err != nil {
			return nil, err
		}
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM([]byte(b))
		if !ok {
			return nil, fmt.Errorf(
				"failed to parse root certificate: %q",
				cfg.caFile,
			)
		}
		if cfg.server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.RootCAs = ca
		}
		tlsConfig.ServerName = cfg.serverAddress
	}
	return tlsConfig, nil
}
