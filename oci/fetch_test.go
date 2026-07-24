// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package oci

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/file"
	ocistore "oras.land/oras-go/v2/content/oci"

	"github.com/defenseunicorns/pkg/helpers/v2"
)

type cachePayload struct {
	Name string `json:"name" yaml:"name"`
}

// warm hits the origin; offline's origin is empty, so any read it serves came from
// the shared cache - returning a value without error proves the path went through src().
func (suite *OCISuite) cachedAndOfflineRemotes(ctx context.Context) (*ocistore.Store, *OrasRemote, *OrasRemote) {
	suite.T().Helper()
	store, err := ocistore.New(suite.T().TempDir())
	suite.NoError(err)
	warm, err := NewOrasRemote("oci://"+suite.remote.Repo().Reference.String(),
		PlatformForArch(testArch), WithPlainHTTP(true), WithCache(store))
	suite.NoError(err)
	offline, err := NewOrasRemote(suite.setupInMemoryRegistry(ctx),
		PlatformForArch(testArch), WithPlainHTTP(true), WithCache(store))
	suite.NoError(err)
	return store, warm, offline
}

func (suite *OCISuite) TestFetchLayerCache() {
	ctx := context.TODO()

	srcTempDir := suite.T().TempDir()
	fileName := "cached-file"
	fileContents := "cache me if you can"
	path := filepath.Join(srcTempDir, fileName)
	suite.NoError(os.WriteFile(path, []byte(fileContents), helpers.ReadWriteUser))
	src, err := file.New(srcTempDir)
	suite.NoError(err)
	desc, err := src.Add(ctx, fileName, ocispec.MediaTypeImageLayer, path)
	suite.NoError(err)
	suite.publishPackage(src, []ocispec.Descriptor{desc})

	store, warm, offline := suite.cachedAndOfflineRemotes(ctx)

	exists, err := store.Exists(ctx, desc)
	suite.NoError(err)
	suite.False(exists)

	b, err := warm.FetchLayer(ctx, desc)
	suite.NoError(err)
	suite.Equal(fileContents, string(b))

	exists, err = store.Exists(ctx, desc)
	suite.NoError(err)
	suite.True(exists) // populated by the fetch

	b, err = offline.FetchLayer(ctx, desc)
	suite.NoError(err)
	suite.Equal(fileContents, string(b)) // served from cache
}

func (suite *OCISuite) TestFetchersUseCache() {
	ctx := context.TODO()

	srcTempDir := suite.T().TempDir()
	writeLayer := func(name, contents string) {
		suite.NoError(os.WriteFile(filepath.Join(srcTempDir, name), []byte(contents), helpers.ReadWriteUser))
	}
	writeLayer("data.json", `{"name":"json-layer"}`)
	writeLayer("data.yaml", "name: yaml-layer\n")

	src, err := file.New(srcTempDir)
	suite.NoError(err)
	var descs []ocispec.Descriptor
	for _, name := range []string{"data.json", "data.yaml"} {
		desc, err := src.Add(ctx, name, ocispec.MediaTypeImageLayer, filepath.Join(srcTempDir, name))
		suite.NoError(err)
		descs = append(descs, desc)
	}
	suite.publishPackage(src, descs)

	rootDesc, err := suite.remote.ResolveRoot(ctx)
	suite.NoError(err)
	root, err := suite.remote.FetchRoot(ctx)
	suite.NoError(err)
	jsonDesc := root.Locate("data.json")

	_, warm, offline := suite.cachedAndOfflineRemotes(ctx)

	// warm each fetcher through the origin, then assert offline returns the same
	// value from cache without returning an error (which it would w/o pre-warming)
	warmM, err := warm.FetchManifest(ctx, rootDesc)
	suite.NoError(err)
	offlineM, err := offline.FetchManifest(ctx, rootDesc)
	suite.NoError(err)
	suite.Equal(warmM, offlineM)

	warmJSON, err := FetchJSONFile[cachePayload](ctx, warm, root, "data.json")
	suite.NoError(err)
	offlineJSON, err := FetchJSONFile[cachePayload](ctx, offline, root, "data.json")
	suite.NoError(err)
	suite.Equal(warmJSON, offlineJSON)

	warmYAML, err := FetchYAMLFile[cachePayload](ctx, warm, root, "data.yaml")
	suite.NoError(err)
	offlineYAML, err := FetchYAMLFile[cachePayload](ctx, offline, root, "data.yaml")
	suite.NoError(err)
	suite.Equal(warmYAML, offlineYAML)

	warmUn, err := FetchUnmarshal[cachePayload](ctx, warm, json.Unmarshal, jsonDesc)
	suite.NoError(err)
	offlineUn, err := FetchUnmarshal[cachePayload](ctx, offline, json.Unmarshal, jsonDesc)
	suite.NoError(err)
	suite.Equal(warmUn, offlineUn)
}
