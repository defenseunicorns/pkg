// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/kstatus/watcher"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func TestWaitForReady(t *testing.T) {
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
	err := WaitForReady(context.Background(), sw, objs)
	require.NoError(t, err)
}

func TestWaitForReadyCanceled(t *testing.T) {
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
	err := WaitForReady(ctx, sw, objs)
	require.EqualError(t, err, "context canceled")
}
