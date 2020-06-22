package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmhodges/clock"
	"github.com/letsencrypt/boulder/cmd"
	"github.com/letsencrypt/boulder/core"
	"github.com/letsencrypt/boulder/features"
	"github.com/letsencrypt/boulder/goodkey"
	bgrpc "github.com/letsencrypt/boulder/grpc"
	blog "github.com/letsencrypt/boulder/log"
	noncepb "github.com/letsencrypt/boulder/nonce/proto"
	rapb "github.com/letsencrypt/boulder/ra/proto"
	sapb "github.com/letsencrypt/boulder/sa/proto"
	"github.com/letsencrypt/boulder/wfe2"
	"github.com/prometheus/client_golang/prometheus"
)

type config struct {
	WFE struct {
		cmd.ServiceConfig
		ListenAddress    string
		TLSListenAddress string

		ServerCertificatePath string
		ServerKeyPath         string

		AllowOrigins []string

		ShutdownStopTimeout cmd.ConfigDuration

		SubscriberAgreementURL string

		TLS cmd.TLSConfig

		RAService *cmd.GRPCClientConfig
		SAService *cmd.GRPCClientConfig
		// GetNonceService contains a gRPC config for any nonce-service instances
		// which we want to retrieve nonces from. In a multi-DC deployment this
		// should refer to any local nonce-service instances.
		GetNonceService *cmd.GRPCClientConfig
		// RedeemNonceServices contains a map of nonce-service prefixes to
		// gRPC configs we want to use to redeem nonces. In a multi-DC deployment
		// this should contain all nonce-services from all DCs as we want to be
		// able to redeem nonces generated at any DC.
		RedeemNonceServices map[string]cmd.GRPCClientConfig

		// CertificateChains maps AIA issuer URLs to certificate filenames.
		// Certificates are read into the chain in the order they are defined in the
		// slice of filenames.
		CertificateChains map[string][]string

		// AlternateCertificateChains maps AIA issuer URLs to an optional alternate
		// certificate chain, represented by an ordered slice of certificate filenames.
		AlternateCertificateChains map[string][]string

		Features map[string]bool

		// DirectoryCAAIdentity is used for the /directory response's "meta"
		// element's "caaIdentities" field. It should match the VA's "issuerDomain"
		// configuration value (this value is the one used to enforce CAA)
		DirectoryCAAIdentity string
		// DirectoryWebsite is used for the /directory response's "meta" element's
		// "website" field.
		DirectoryWebsite string

		// ACMEv2 requests (outside some registration/revocation messages) use a JWS with
		// a KeyID header containing the full account URL. For new accounts this
		// will be a KeyID based on the HTTP request's Host header and the ACMEv2
		// account path. For legacy ACMEv1 accounts we need to whitelist the account
		// ID prefix that legacy accounts would have been using based on the Host
		// header of the WFE1 instance and the legacy 'reg' path component. This
		// will differ in configuration for production and staging.
		LegacyKeyIDPrefix string

		// BlockedKeyFile is the path to a YAML file containing Base64 encoded
		// SHA256 hashes of SubjectPublicKeyInfo's that should be considered
		// administratively blocked.
		BlockedKeyFile string

		// StaleTimeout determines how old should data be to be accessed via Boulder-specific GET-able APIs
		StaleTimeout cmd.ConfigDuration

		// AuthorizationLifetimeDays defines how long authorizations will be
		// considered valid for. The WFE uses this to find the creation date of
		// authorizations by subtracing this value from the expiry. It should match
		// the value configured in the RA.
		AuthorizationLifetimeDays int

		// PendingAuthorizationLifetimeDays defines how long authorizations may be in
		// the pending state before expiry. The WFE uses this to find the creation
		// date of pending authorizations by subtracting this value from the expiry.
		// It should match the value configured in the RA.
		PendingAuthorizationLifetimeDays int
	}

	Syslog cmd.SyslogConfig

	Common struct {
		IssuerCert string
	}
}

// loadCertificateFile loads a PEM certificate from the certFile provided. It
// validates that the PEM is well-formed with no leftover bytes, and contains
// only a well-formed X509 certificate. If the cert file meets these
// requirements the PEM bytes from the file are returned along with the parsed
// certificate, otherwise an error is returned. If the PEM contents of
// a certFile do not have a trailing newline one is added.
func loadCertificateFile(aiaIssuerURL, certFile string) ([]byte, *x509.Certificate, error) {
	pemBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"CertificateChain entry for AIA issuer url %q has an "+
				"invalid chain file: %q - error reading contents: %s",
			aiaIssuerURL, certFile, err)
	}
	if bytes.Contains(pemBytes, []byte("\r\n")) {
		return nil, nil, fmt.Errorf(
			"CertificateChain entry for AIA issuer url %q has an "+
				"invalid chain file: %q - contents had CRLF line endings",
			aiaIssuerURL, certFile)
	}
	// Try to decode the contents as PEM
	certBlock, rest := pem.Decode(pemBytes)
	if certBlock == nil {
		return nil, nil, fmt.Errorf(
			"CertificateChain entry for AIA issuer url %q has an "+
				"invalid chain file: %q - contents did not decode as PEM",
			aiaIssuerURL, certFile)
	}
	// The PEM contents must be a CERTIFICATE
	if certBlock.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf(
			"CertificateChain entry for AIA issuer url %q has an "+
				"invalid chain file: %q - PEM block type incorrect, found "+
				"%q, expected \"CERTIFICATE\"",
			aiaIssuerURL, certFile, certBlock.Type)
	}
	// The PEM Certificate must successfully parse
	var cert *x509.Certificate
	if cert, err = x509.ParseCertificate(certBlock.Bytes); err != nil {
		return nil, nil, fmt.Errorf(
			"CertificateChain entry for AIA issuer url %q has an "+
				"invalid chain file: %q - certificate bytes failed to parse: %s",
			aiaIssuerURL, certFile, err)
	}
	// If there are bytes leftover we must reject the file otherwise these
	// leftover bytes will end up in a served certificate chain.
	if len(rest) != 0 {
		return nil, nil, fmt.Errorf(
			"CertificateChain entry for AIA issuer url %q has an "+
				"invalid chain file: %q - PEM contents had unused remainder "+
				"input (%d bytes)",
			aiaIssuerURL, certFile, len(rest))
	}
	// If the PEM contents don't end in a \n, add it.
	if pemBytes[len(pemBytes)-1] != '\n' {
		pemBytes = append(pemBytes, '\n')
	}
	return pemBytes, cert, nil
}

// loadCertificateChains processes the provided chainConfig of AIA Issuer URLs
// and cert filenames. For each AIA issuer URL all of its cert filenames are
// read, validated as PEM certificates, and concatenated together separated by
// newlines. The combined PEM certificate chain contents for each are returned
// in the results map, keyed by the AIA Issuer URL. Additionally the first
// certificate in each chain is parsed and returned in a slice of issuer
// certificates.
func loadCertificateChains(chainConfig map[string][]string, requireAtLeastOneChain bool) (map[string][]byte, []*x509.Certificate, error) {
	results := make(map[string][]byte, len(chainConfig))
	var issuerCerts []*x509.Certificate

	// For each AIA Issuer URL we need to read the chain cert files
	for aiaIssuerURL, certFiles := range chainConfig {
		var buffer bytes.Buffer

		// There must be at least one chain file specified
		if requireAtLeastOneChain && len(certFiles) == 0 {
			return nil, nil, fmt.Errorf(
				"CertificateChain entry for AIA issuer url %q has no chain "+
					"file names configured",
				aiaIssuerURL)
		}

		// certFiles are read and appended in the order they appear in the
		// configuration
		for i, c := range certFiles {
			// Prepend a newline before each chain entry
			buffer.Write([]byte("\n"))

			// Read and validate the chain file contents
			pemBytes, cert, err := loadCertificateFile(aiaIssuerURL, c)
			if err != nil {
				return nil, nil, err
			}

			// Save the first certificate as a direct issuer certificate
			if i == 0 {
				issuerCerts = append(issuerCerts, cert)
			}

			// Write the PEM bytes to the result buffer for this AIAIssuer
			buffer.Write(pemBytes)
		}

		// Save the full PEM chain contents, if any
		if buffer.Len() > 0 {
			results[aiaIssuerURL] = buffer.Bytes()
		}
	}

	return results, issuerCerts, nil
}

func setupWFE(c config, logger blog.Logger, stats prometheus.Registerer, clk clock.Clock) (core.RegistrationAuthority, core.StorageAuthority, noncepb.NonceServiceClient, map[string]noncepb.NonceServiceClient) {
	tlsConfig, err := c.WFE.TLS.Load()
	cmd.FailOnError(err, "TLS config")
	clientMetrics := bgrpc.NewClientMetrics(stats)
	raConn, err := bgrpc.ClientSetup(c.WFE.RAService, tlsConfig, clientMetrics, clk)
	cmd.FailOnError(err, "Failed to load credentials and create gRPC connection to RA")
	rac := bgrpc.NewRegistrationAuthorityClient(rapb.NewRegistrationAuthorityClient(raConn))

	saConn, err := bgrpc.ClientSetup(c.WFE.SAService, tlsConfig, clientMetrics, clk)
	cmd.FailOnError(err, "Failed to load credentials and create gRPC connection to SA")
	sac := bgrpc.NewStorageAuthorityClient(sapb.NewStorageAuthorityClient(saConn))

	var rns noncepb.NonceServiceClient
	npm := map[string]noncepb.NonceServiceClient{}
	if c.WFE.GetNonceService != nil {
		rnsConn, err := bgrpc.ClientSetup(c.WFE.GetNonceService, tlsConfig, clientMetrics, clk)
		cmd.FailOnError(err, "Failed to load credentials and create gRPC connection to get nonce service")
		rns = noncepb.NewNonceServiceClient(rnsConn)
		for prefix, serviceConfig := range c.WFE.RedeemNonceServices {
			conn, err := bgrpc.ClientSetup(&serviceConfig, tlsConfig, clientMetrics, clk)
			cmd.FailOnError(err, "Failed to load credentials and create gRPC connection to redeem nonce service")
			npm[prefix] = noncepb.NewNonceServiceClient(conn)
		}
	}

	return rac, sac, rns, npm
}

type errorWriter struct {
	blog.Logger
}

func (ew errorWriter) Write(p []byte) (n int, err error) {
	// log.Logger will append a newline to all messages before calling
	// Write. Our log checksum checker doesn't like newlines, because
	// syslog will strip them out so the calculated checksums will
	// differ. So that we don't hit this corner case for every line
	// logged from inside net/http.Server we strip the newline before
	// we get to the checksum generator.
	p = bytes.TrimRight(p, "\n")
	ew.Logger.Err(fmt.Sprintf("net/http.Server: %s", string(p)))
	return
}

func main() {
	configFile := flag.String("config", "", "File path to the configuration file for this service")
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	var c config
	err := cmd.ReadConfigFile(*configFile, &c)
	cmd.FailOnError(err, "Reading JSON config file into config structure")

	// Map of AIA Issuer URLs to a slice of PEM-encoded certificate chains.
	// The first chain in the slice is the default chain, and subsequent
	// chains are alternates.
	allCertChains := make(map[string][][]byte, len(c.WFE.CertificateChains))

	certChains, issuerCerts, err := loadCertificateChains(c.WFE.CertificateChains, true)
	cmd.FailOnError(err, "Couldn't read configured CertificateChains")

	for aiaURL, chainPEM := range certChains {
		allCertChains[aiaURL] = [][]byte{chainPEM}
	}

	err = features.Set(c.WFE.Features)
	cmd.FailOnError(err, "Failed to set feature flags")

	if c.WFE.AlternateCertificateChains != nil {
		altCertChains, _, err := loadCertificateChains(c.WFE.AlternateCertificateChains, false)
		cmd.FailOnError(err, "Couldn't read configured AlternateCertificateChains")

		for aiaURL, chainPEM := range altCertChains {
			if _, ok := allCertChains[aiaURL]; !ok {
				cmd.Fail(fmt.Sprintf("AIA Issuer URL %s appeared in AlternateCertificateChains, "+
					"but does not exist in CertificateChains", aiaURL))
			}
			allCertChains[aiaURL] = append(allCertChains[aiaURL], chainPEM)
		}
	}

	stats, logger := cmd.StatsAndLogging(c.Syslog, c.WFE.DebugAddr)
	defer logger.AuditPanic()
	logger.Info(cmd.VersionString())

	clk := cmd.Clock()

	rac, sac, rns, npm := setupWFE(c, logger, stats, clk)
	// don't load any weak keys, but do load blocked keys
	kp, err := goodkey.NewKeyPolicy("", c.WFE.BlockedKeyFile, sac.KeyBlocked)
	cmd.FailOnError(err, "Unable to create key policy")

	if c.WFE.StaleTimeout.Duration == 0 {
		c.WFE.StaleTimeout.Duration = time.Minute * 10
	}

	authorizationLifetime := 30 * (24 * time.Hour)
	if c.WFE.AuthorizationLifetimeDays != 0 {
		authorizationLifetime = time.Duration(c.WFE.AuthorizationLifetimeDays) * (24 * time.Hour)
	}

	pendingAuthorizationLifetime := 7 * (24 * time.Hour)
	if c.WFE.PendingAuthorizationLifetimeDays != 0 {
		pendingAuthorizationLifetime = time.Duration(c.WFE.PendingAuthorizationLifetimeDays) * (24 * time.Hour)
	}

	wfe, err := wfe2.NewWebFrontEndImpl(stats, clk, kp, allCertChains, issuerCerts, rns, npm, logger, c.WFE.StaleTimeout.Duration, authorizationLifetime, pendingAuthorizationLifetime)
	cmd.FailOnError(err, "Unable to create WFE")
	wfe.RA = rac
	wfe.SA = sac

	wfe.SubscriberAgreementURL = c.WFE.SubscriberAgreementURL
	wfe.AllowOrigins = c.WFE.AllowOrigins
	wfe.DirectoryCAAIdentity = c.WFE.DirectoryCAAIdentity
	wfe.DirectoryWebsite = c.WFE.DirectoryWebsite
	wfe.LegacyKeyIDPrefix = c.WFE.LegacyKeyIDPrefix

	wfe.IssuerCert, err = cmd.LoadCert(c.Common.IssuerCert)
	cmd.FailOnError(err, fmt.Sprintf("Couldn't read issuer cert [%s]", c.Common.IssuerCert))

	logger.Infof("WFE using key policy: %#v", kp)

	logger.Infof("Server running, listening on %s....", c.WFE.ListenAddress)
	handler := wfe.Handler(stats)
	srv := http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         c.WFE.ListenAddress,
		ErrorLog:     log.New(errorWriter{logger}, "", 0),
		Handler:      handler,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			cmd.FailOnError(err, "Running HTTP server")
		}
	}()

	tlsSrv := http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         c.WFE.TLSListenAddress,
		ErrorLog:     log.New(errorWriter{logger}, "", 0),
		Handler:      handler,
	}
	if tlsSrv.Addr != "" {
		go func() {
			err := tlsSrv.ListenAndServeTLS(c.WFE.ServerCertificatePath, c.WFE.ServerKeyPath)
			if err != nil && err != http.ErrServerClosed {
				cmd.FailOnError(err, "Running TLS server")
			}
		}()
	}

	done := make(chan bool)
	go cmd.CatchSignals(logger, func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.WFE.ShutdownStopTimeout.Duration)
		defer cancel()
		_ = srv.Shutdown(ctx)
		_ = tlsSrv.Shutdown(ctx)
		done <- true
	})

	// https://godoc.org/net/http#Server.Shutdown:
	// When Shutdown is called, Serve, ListenAndServe, and ListenAndServeTLS
	// immediately return ErrServerClosed. Make sure the program doesn't exit and
	// waits instead for Shutdown to return.
	<-done
}
