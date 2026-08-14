// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package oci

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func TestNewOrasRemote_TransportIsolation(t *testing.T) {
	retryTransport := retry.DefaultClient.Transport
	platform := PlatformForArch(testArch)

	first, err := NewOrasRemote("example.com/first:latest", platform)
	require.NoError(t, err)
	second, err := NewOrasRemote("example.com/second:latest", platform)
	require.NoError(t, err)

	firstClient, ok := first.repo.Client.(*auth.Client)
	require.True(t, ok)
	secondClient, ok := second.repo.Client.(*auth.Client)
	require.True(t, ok)
	require.NotSame(t, firstClient.Client, secondClient.Client)
	require.NotSame(t, firstClient.Cache, secondClient.Cache)
	require.NotSame(t, first.progTransport, second.progTransport)
	require.NotSame(t, first.progTransport.Base, second.progTransport.Base)
	require.Same(t, retryTransport, retry.DefaultClient.Transport)
}

func TestWithTransport_SurvivesProgressChanges(t *testing.T) {
	transport := &http.Transport{}
	remote, err := NewOrasRemote(
		"example.com/repository:latest",
		PlatformForArch(testArch),
		WithTransport(transport),
	)
	require.NoError(t, err)
	require.NotSame(t, transport, remote.progTransport.Base)
	configuredTransport, ok := remote.progTransport.Base.(*http.Transport)
	require.True(t, ok)

	client, ok := remote.repo.Client.(*auth.Client)
	require.True(t, ok)
	require.Same(t, remote.progTransport, client.Client.Transport)

	remote.SetProgressWriter(&TestProgressWriter{})
	require.Same(t, configuredTransport, remote.progTransport.Base)
	require.Same(t, remote.progTransport, client.Client.Transport)

	remote.ClearProgressWriter()
	require.Same(t, configuredTransport, remote.progTransport.Base)
	require.Same(t, remote.progTransport, client.Client.Transport)
}

func TestWithTransport_UsedForRequests(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	transport, ok := server.Client().Transport.(*http.Transport)
	require.True(t, ok)
	remote, err := NewOrasRemote(
		"example.com/repository:latest",
		PlatformForArch(testArch),
		WithTransport(transport),
	)
	require.NoError(t, err)

	client, ok := remote.repo.Client.(*auth.Client)
	require.True(t, ok)
	response, err := client.Client.Get(server.URL)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestWithUserAgent_SurvivesRepositorySetup(t *testing.T) {
	remote, err := NewOrasRemote(
		"example.com/repository:latest",
		PlatformForArch(testArch),
		WithUserAgent("zarf/test"),
	)
	require.NoError(t, err)

	client, ok := remote.repo.Client.(*auth.Client)
	require.True(t, ok)
	require.Equal(t, "zarf/test", client.Header.Get("User-Agent"))
	require.NotNil(t, client.Credential)
}

func TestWithInsecureSkipVerify_ClonesTransport(t *testing.T) {
	tests := []struct {
		name      string
		tlsConfig *tls.Config
	}{
		{name: "nil TLS config"},
		{name: "existing TLS config", tlsConfig: &tls.Config{ServerName: "registry.example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &http.Transport{TLSClientConfig: tt.tlsConfig}
			remote, err := NewOrasRemote(
				"example.com/repository:latest",
				PlatformForArch(testArch),
				WithTransport(transport),
				WithInsecureSkipVerify(true),
			)
			require.NoError(t, err)

			configured, ok := remote.progTransport.Base.(*http.Transport)
			require.True(t, ok)
			require.NotSame(t, transport, configured)
			require.NotNil(t, configured.TLSClientConfig)
			require.True(t, configured.TLSClientConfig.InsecureSkipVerify)
			require.False(t, transport.TLSClientConfig.InsecureSkipVerify)
			if tt.tlsConfig != nil {
				require.NotSame(t, tt.tlsConfig, configured.TLSClientConfig)
				require.Equal(t, tt.tlsConfig.ServerName, configured.TLSClientConfig.ServerName)
				require.False(t, tt.tlsConfig.InsecureSkipVerify)
			}
		})
	}
}

func TestWithInsecureSkipVerify_OrderIndependent(t *testing.T) {
	tests := []struct {
		name          string
		insecure      bool
		insecureFirst bool
	}{
		{name: "enable after transport", insecure: true, insecureFirst: false},
		{name: "enable before transport", insecure: true, insecureFirst: true},
		{name: "disable after transport", insecure: false, insecureFirst: false},
		{name: "disable before transport", insecure: false, insecureFirst: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !tt.insecure}}
			mods := []Modifier{WithTransport(transport), WithInsecureSkipVerify(tt.insecure)}
			if tt.insecureFirst {
				mods[0], mods[1] = mods[1], mods[0]
			}

			remote, err := NewOrasRemote(
				"example.com/repository:latest",
				PlatformForArch(testArch),
				mods...,
			)
			require.NoError(t, err)

			configured, ok := remote.progTransport.Base.(*http.Transport)
			require.True(t, ok)
			require.Equal(t, tt.insecure, configured.TLSClientConfig.InsecureSkipVerify)
			require.Equal(t, !tt.insecure, transport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}

func TestWithTransport_PreservesInsecureSkipVerifyWhenUnset(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
	}{
		{name: "secure", insecure: false},
		{name: "insecure", insecure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: tt.insecure}}
			remote, err := NewOrasRemote(
				"example.com/repository:latest",
				PlatformForArch(testArch),
				WithTransport(transport),
			)
			require.NoError(t, err)

			configured, ok := remote.progTransport.Base.(*http.Transport)
			require.True(t, ok)
			require.Equal(t, tt.insecure, configured.TLSClientConfig.InsecureSkipVerify)
		})
	}
}
