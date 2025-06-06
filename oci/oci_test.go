// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package oci

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory" // used for docker test registry
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"

	"github.com/defenseunicorns/pkg/helpers/v2"
)

type OCISuite struct {
	suite.Suite
	*require.Assertions
	remote *OrasRemote
}

func (suite *OCISuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())
	ctx := context.TODO()
	registry := suite.setupInMemoryRegistry(ctx)
	platform := PlatformForArch(testArch)

	var err error
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	suite.remote, err = NewOrasRemote(registry, platform, WithPlainHTTP(true), WithLogger(logger))
	suite.NoError(err)
}

const (
	testArch = "fake-test-arch"
)

func (suite *OCISuite) setupInMemoryRegistry(ctx context.Context) string {
	suite.T().Helper()

	port, err := freeport.GetFreePort()
	suite.NoError(err)
	config := &configuration.Configuration{}
	config.HTTP.Addr = fmt.Sprintf(":%d", port)
	config.HTTP.Secret = "Fake secret so we don't get warning"
	config.Log.AccessLog.Disabled = true
	config.HTTP.DrainTimeout = 10 * time.Second
	config.Storage = map[string]configuration.Parameters{"inmemory": map[string]any{}}

	reg, err := registry.NewRegistry(ctx, config)
	suite.NoError(err)

	go func() {
		_ = reg.ListenAndServe()
	}()

	url := fmt.Sprintf("localhost:%d", port)

	return fmt.Sprintf("oci://%s/package:1.0.1", url)
}

func (suite *OCISuite) TestPublishFailNoTitle() {
	ctx := context.TODO()
	annotations := map[string]string{
		ocispec.AnnotationDescription: "No title",
	}
	_, err := suite.remote.CreateAndPushManifestConfig(ctx, annotations, ocispec.MediaTypeImageConfig)
	suite.Error(err)
}

func (suite *OCISuite) publishPackage(src *file.Store, descs []ocispec.Descriptor) {
	suite.T().Helper()
	ctx := context.TODO()
	annotations := map[string]string{
		ocispec.AnnotationTitle:       "name",
		ocispec.AnnotationDescription: "description",
	}

	manifestConfigDesc, err := suite.remote.CreateAndPushManifestConfig(ctx, annotations, ocispec.MediaTypeLayoutHeader)
	suite.NoError(err)

	manifestDesc, err := suite.remote.PackAndTagManifest(ctx, src, descs, manifestConfigDesc, annotations)
	suite.NoError(err)
	publishedDesc, err := oras.Copy(ctx, src, manifestDesc.Digest.String(), suite.remote.Repo(), "", suite.remote.GetDefaultCopyOpts())
	suite.NoError(err)

	err = suite.remote.UpdateIndex(ctx, "0.0.1", publishedDesc)
	suite.NoError(err)
}

func (suite *OCISuite) TestCopyToTarget() {
	ctx := context.TODO()

	srcTempDir := suite.T().TempDir()
	regularFileName := "this-file-is-in-a-regular-directory"
	fileContents := "here's what I'm putting in the file"
	ociFileName := "this-file-is-in-a-oci-file-store"

	regularFilePath := filepath.Join(srcTempDir, regularFileName)
	err := os.WriteFile(regularFilePath, []byte(fileContents), helpers.ReadWriteUser)
	suite.NoError(err)
	src, err := file.New(srcTempDir)
	suite.NoError(err)

	desc, err := src.Add(ctx, ociFileName, ocispec.MediaTypeImageLayer, regularFilePath)
	suite.NoError(err)

	descs := []ocispec.Descriptor{desc}
	suite.publishPackage(src, descs)

	otherTempDir := suite.T().TempDir()

	dst, err := file.New(otherTempDir)
	suite.NoError(err)

	err = suite.remote.CopyToTarget(ctx, descs, dst, suite.remote.GetDefaultCopyOpts())
	suite.NoError(err)

	ociFile := filepath.Join(otherTempDir, ociFileName)
	b, err := os.ReadFile(ociFile)
	suite.NoError(err)
	contents := string(b)
	suite.Equal(contents, fileContents)
}

func (suite *OCISuite) TestPulledPaths() {
	ctx := context.TODO()
	srcTempDir := suite.T().TempDir()
	err := helpers.CreateDirectory(filepath.Join(srcTempDir, "subdir"), helpers.ReadExecuteAllWriteUser)
	suite.NoError(err)
	files := []string{"firstFile", filepath.Join("subdir", "secondFile")}

	var descs []ocispec.Descriptor
	src, err := file.New(srcTempDir)
	suite.NoError(err)
	for _, file := range files {
		path := filepath.Join(srcTempDir, file)
		f, err := os.Create(path)
		suite.NoError(err)
		defer f.Close()
		desc, err := src.Add(ctx, file, ocispec.MediaTypeEmptyJSON, path)
		suite.NoError(err)
		descs = append(descs, desc)
	}

	suite.publishPackage(src, descs)
	dstTempDir := suite.T().TempDir()

	_, err = suite.remote.PullPaths(ctx, dstTempDir, files)
	suite.NoError(err)
	for _, file := range files {
		pulledPathOCIFile := filepath.Join(dstTempDir, file)
		_, err := os.Stat(pulledPathOCIFile)
		suite.NoError(err)
	}
}

func (suite *OCISuite) TestResolveRoot() {
	suite.T().Log("Testing resolve root")
	ctx := context.TODO()
	srcTempDir := suite.T().TempDir()
	files := []string{"ResolveRootFile1", "ResolveRootFile2", "ResolveRootFile3"}

	var descs []ocispec.Descriptor
	src, err := file.New(srcTempDir)
	suite.NoError(err)
	for _, file := range files {
		path := filepath.Join(srcTempDir, file)
		f, err := os.Create(path)
		suite.NoError(err)
		defer f.Close()
		desc, err := src.Add(ctx, file, ocispec.MediaTypeEmptyJSON, path)
		suite.NoError(err)
		descs = append(descs, desc)
	}

	suite.publishPackage(src, descs)

	root, err := suite.remote.FetchRoot(ctx)
	suite.NoError(err)
	suite.Len(root.Layers, 3)
	desc := root.Locate("ResolveRootFile3")
	suite.Equal("ResolveRootFile3", desc.Annotations[ocispec.AnnotationTitle])
}

func (tpw *TestProgressWriter) Write(b []byte) (int, error) {
	tpw.bytesSent += len(b)
	return len(b), nil
}

func (TestProgressWriter) Close() error {
	return nil
}

func (TestProgressWriter) Updatef(s string, a ...any) {
	fmt.Printf(s, a...)
}

func (TestProgressWriter) Successf(s string, a ...any) {
	fmt.Printf(s, a...)
}

func (TestProgressWriter) Failf(s string, a ...any) {
	fmt.Printf(s, a...)
}

type TestProgressWriter struct {
	bytesSent int
}

func (suite *OCISuite) TestCopy() {
	suite.T().Log("Testing copying between OCI remotes")
	ctx := context.TODO()
	srcTempDir := suite.T().TempDir()
	files := []string{"firstFile", "secondFile"}

	fileContents := "here's what I'm putting in each file"

	var descs []ocispec.Descriptor
	src, err := file.New(srcTempDir)
	suite.NoError(err)
	for _, file := range files {
		path := filepath.Join(srcTempDir, file)
		err := os.WriteFile(path, []byte(fileContents), helpers.ReadWriteUser)
		suite.NoError(err)
		desc, err := src.Add(ctx, file, ocispec.MediaTypeImageLayer, path)
		suite.NoError(err)
		descs = append(descs, desc)
	}

	suite.publishPackage(src, descs)

	dstRegistryURL := suite.setupInMemoryRegistry(ctx)
	dstRemote, err := NewOrasRemote(dstRegistryURL, PlatformForArch(testArch), WithPlainHTTP(true))
	suite.NoError(err)
	testWriter := &TestProgressWriter{}
	err = Copy(ctx, suite.remote, dstRemote, nil, 1, testWriter)
	suite.NoError(err)

	srcRoot, err := suite.remote.FetchRoot(ctx)
	suite.NoError(err)
	totalSize := srcRoot.Config.Size
	for _, layer := range srcRoot.Layers {
		totalSize += layer.Size
		ok, err := dstRemote.Repo().Exists(ctx, layer)
		suite.True(ok)
		suite.NoError(err)
	}

	suite.Equal(int(totalSize), testWriter.bytesSent)
}

func TestRemoveDuplicateDescriptors(t *testing.T) {
	tests := []struct {
		name     string
		input    []ocispec.Descriptor
		expected []ocispec.Descriptor
	}{
		{
			name: "no duplicates",
			input: []ocispec.Descriptor{
				{Digest: "sha256:1111", Size: 100},
				{Digest: "sha256:2222", Size: 200},
			},
			expected: []ocispec.Descriptor{
				{Digest: "sha256:1111", Size: 100},
				{Digest: "sha256:2222", Size: 200},
			},
		},
		{
			name: "with duplicates",
			input: []ocispec.Descriptor{
				{Digest: "sha256:1111", Size: 100},
				{Digest: "sha256:1111", Size: 100},
				{Digest: "sha256:2222", Size: 200},
			},
			expected: []ocispec.Descriptor{
				{Digest: "sha256:1111", Size: 100},
				{Digest: "sha256:2222", Size: 200},
			},
		},
		{
			name: "with empty descriptor",
			input: []ocispec.Descriptor{
				{Size: 0},
				{Digest: "sha256:1111", Size: 100},
			},
			expected: []ocispec.Descriptor{
				{Digest: "sha256:1111", Size: 100},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicateDescriptors(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveDuplicateDescriptors(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOCI(t *testing.T) {
	suite.Run(t, new(OCISuite))
}
