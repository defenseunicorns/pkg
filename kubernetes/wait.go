// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/aggregator"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/collector"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/kstatus/watcher"
	"sigs.k8s.io/cli-utils/pkg/object"
)

// WatcherForConfig returns a status watcher for the give Kubernetes configuration.
func WatcherForConfig(cfg *rest.Config) (watcher.StatusWatcher, error) {
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)
	sw := watcher.NewDefaultStatusWatcher(dynamicClient, rm)
	return sw, nil
}

// WaitForReadyRuntime waits for all of the runtime objects to reach a ready state.
func WaitForReadyRuntime(ctx context.Context, sw watcher.StatusWatcher, robjs []runtime.Object) error {
	objs := []object.ObjMetadata{}
	for _, robj := range robjs {
		obj, err := object.RuntimeToObjMeta(robj)
		if err != nil {
			return err
		}
		objs = append(objs, obj)
	}
	return WaitForReady(ctx, sw, objs)
}

// WaitForReady waits for all of the objects to reach a ready state.
func WaitForReady(ctx context.Context, sw watcher.StatusWatcher, objs []object.ObjMetadata) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	eventCh := sw.Watch(cancelCtx, objs, watcher.Options{})
	statusCollector := collector.NewResourceStatusCollector(objs)
	done := statusCollector.ListenWithObserver(eventCh, collector.ObserverFunc(
		func(statusCollector *collector.ResourceStatusCollector, _ event.Event) {
			rss := []*event.ResourceStatus{}
			for _, rs := range statusCollector.ResourceStatuses {
				if rs == nil {
					continue
				}
				rss = append(rss, rs)
			}
			desired := status.CurrentStatus
			if aggregator.AggregateStatus(rss, desired) == desired {
				cancel()
				return
			}
		}),
	)
	<-done
	if statusCollector.Error != nil {
		return statusCollector.Error
	}
	// Only check parent context error, otherwise we would error when desired status is acheived.
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}
