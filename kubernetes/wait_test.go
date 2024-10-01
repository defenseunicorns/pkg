// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package kubernetes

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/kstatus/watcher"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func TestWaitForReady(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sw := NewImmediateWatcher(status.CurrentStatus)
	objs := []object.ObjMetadata{
		{
			GroupKind: schema.GroupKind{
				Group: "apps",
				Kind:  "Deployment",
			},
			Namespace: "foo",
			Name:      "bar",
		},
	}
	err := WaitForReady(context.Background(), sw, objs, logger)
	require.NoError(t, err)
	logOutput := buf.String()
	require.Contains(t, logOutput, "bar: deployment ready")
}

func TestWaitForReadyCanceled(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sw := watcher.BlindStatusWatcher{}
	objs := []object.ObjMetadata{
		{
			GroupKind: schema.GroupKind{
				Group: "apps",
				Kind:  "Deployment",
			},
			Namespace: "foo",
			Name:      "bar",
		},
	}
	err := WaitForReady(ctx, sw, objs, logger)
	require.EqualError(t, err, "context canceled")
	logOutput := buf.String()
	require.Contains(t, logOutput, "bar: deployment not ready")
}
