// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package kubernetes

import (
	"context"

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

// WaitForReady waits for all of the resources to reach a ready state.
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

// ImmediateWatcher should only be used for testing and returns the set status immediatly.
type ImmediateWatcher struct {
	status status.Status
}

// NewImmediateWatcher returns a ImmediateWatcher.
func NewImmediateWatcher(status status.Status) *ImmediateWatcher {
	return &ImmediateWatcher{
		status: status,
	}
}

// Watch watches the given objects and immediatly returns the configured status.
func (w *ImmediateWatcher) Watch(_ context.Context, objs object.ObjMetadataSet, _ watcher.Options) <-chan event.Event {
	eventCh := make(chan event.Event, len(objs))
	for _, obj := range objs {
		eventCh <- event.Event{
			Type: event.ResourceUpdateEvent,
			Resource: &event.ResourceStatus{
				Identifier: obj,
				Status:     w.status,
			},
		}
	}
	close(eventCh)
	return eventCh
}
