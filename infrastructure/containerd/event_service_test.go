package containerd_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/weaveworks/reignite/api/events"
	"github.com/weaveworks/reignite/core/ports"
	"github.com/weaveworks/reignite/infrastructure/containerd"
)

func TestEventService_Integration(t *testing.T) {
	if !runContainerDTests() {
		t.Skip("skipping containerd event service integration test")
	}

	RegisterTestingT(t)

	client, ctx := testCreateClient(t)

	es := containerd.NewEventServiceWithClient(&containerd.Config{
		SnapshotterKernel: testSnapshotter,
		SnapshotterVolume: testSnapshotter,
		Namespace:         testContainerdNs,
	}, client)

	t.Log("creating subscribers")

	ctx1, cancel1 := context.WithCancel(ctx)
	evt1, err1 := es.Subscribe(ctx1)
	ctx2, cancel2 := context.WithCancel(ctx)
	evt2, err2 := es.Subscribe(ctx2)

	errChan := make(chan error)

	testEvents := []*events.MicroVMSpecCreated{
		{ID: "vm1", Namespace: "ns1"},
		{ID: "vm2", Namespace: "ns1"},
	}

	subscribers := []testSubscriber{
		{eventCh: evt1, eventErrCh: err1, cancel: cancelWrapper(t, 1, cancel1)},
		{eventCh: evt2, eventErrCh: err2, cancel: cancelWrapper(t, 2, cancel2)},
	}

	go func() {
		defer close(errChan)

		for _, event := range testEvents {
			if err := es.Publish(ctx, "/reignite/test", event); err != nil {
				errChan <- err
				return
			}
		}

		t.Log("finished publishing events")
	}()

	t.Log("subscribers waiting for events")
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}

	for idx, subscriber := range subscribers {
		t.Logf("start subscriber (%d) is ready to receive events", idx+1)
		recvd, err := watch(&subscriber, len(testEvents))
		t.Logf("subscriber (%d) is done", idx+1)

		assert.NoError(t, err)
		assert.Len(t, recvd, 2)

		//		if len(recvd) == len(testEvents) {
		//			subscriber.cancel()
		//		}
	}
}

func cancelWrapper(t *testing.T, id int, cancel context.CancelFunc) func() {
	return func() {
		t.Logf("context (%d) cancelled", id)
		cancel()
	}
}

func watch(subscriber *testSubscriber, maxEvents int) ([]interface{}, error) {
	recvd := []interface{}{}

	var err error

	for {
		select {
		case env := <-subscriber.eventCh:
			if env != nil {
				recvd = append(recvd, env.Event)
			} else {
				break
			}
		case err = <-subscriber.eventErrCh:
			break
		}

		if len(recvd) == maxEvents {
			subscriber.cancel()
			break
		}
	}

	return recvd, err
}

type testEvent struct {
	Name  string
	Value string
}

type testSubscriber struct {
	eventCh    <-chan *ports.EventEnvelope
	eventErrCh <-chan error
	cancel     func()
}
