package main

import (
	"log"

	fpostgresstore "github.com/idirall22/twee/follow/store/postgres"

	"github.com/idirall22/twee/follow"

	"github.com/idirall22/twee/common"
	"github.com/idirall22/twee/notification"
	neventstore "github.com/idirall22/twee/notification/event_store/stan"
	nstore "github.com/idirall22/twee/notification/store/postgres"
)

func main() {
	ns, err := nstore.NewPostgresNotificationStore(common.PostgresTestOptions)
	if err != nil {
		log.Fatalf("Could not create notification store: %v", err)
	}

	fns, err := fpostgresstore.NewPostgresFollowStore(common.PostgresTestOptions)
	if err != nil {
		log.Fatalf("Could not create notification store: %v", err)
	}

	fs, err := follow.NewFollowServer(fns, nil)
	if err != nil {
		log.Fatalf("Could not connect to event store: %v", err)
	}

	es, err := neventstore.NewNatsStreamingEventStore(
		"tweets",
		"test-cluster",
		"test-cluster-01",
		ns,
		fs,
	)

	if err != nil {
		log.Fatalf("Could not connect to event store: %v", err)
	}

	service, err := notification.NewNotificationServer(ns, es, fs)
	if err != nil {
		log.Fatal(err)
	}
	defer service.Close()
	service.Start()
}
