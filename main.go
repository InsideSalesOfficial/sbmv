package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	servicebus "github.com/Azure/azure-service-bus-go"
)

func main() {
	topicName := flag.String("topic", "", "topic name")
	subscriptionName := flag.String("subscription", "", "subscription name")
	flag.Parse()
	ctx := context.Background()

	if *topicName == "" || *subscriptionName == "" {
		flag.Usage()
		os.Exit(1)
	}

	connStr := os.Getenv("SERVICEBUS_CONNECTION_STRING")
	if connStr == "" {
		log.Fatal("SERVICEBUS_CONNECTION_STRING required")
	}

	log.Printf("Topic: %v Subscription: %v", *topicName, *subscriptionName)

	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connStr))
	if err != nil {
		log.Fatal(err)
	}

	tm := ns.NewTopicManager()
	topicEntity, err := ensureTopic(ctx, tm, *topicName)
	if err != nil {
		log.Fatal(err)
	}

	topic, err := ns.NewTopic(topicEntity.Name)
	if err != nil {
		return
	}

	sm, err := ns.NewSubscriptionManager(topicEntity.Name)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = ensureSubscription(ctx, sm, *subscriptionName)
	if err != nil {
		return
	}

	sub, err := topic.NewSubscription(*subscriptionName)
	if err != nil {
		return
	}

	deadEntity, err := sm.Get(ctx, sub.Name)
	if err != nil {
		log.Fatal("asdf  a ", err)
	}

	deadletterCount := int(*deadEntity.SubscriptionDescription.CountDetails.DeadLetterMessageCount)
	log.Println("Subscription Count Details: ", JSONStringPretty(deadEntity.SubscriptionDescription.CountDetails))
	log.Printf("Found %d dead letter messages, processing...", *deadEntity.SubscriptionDescription.CountDetails.DeadLetterMessageCount)

	qdl := sub.NewDeadLetter()

	for count := 0; count < deadletterCount; count++ {
		err := qdl.ReceiveOne(ctx, servicebus.HandlerFunc(func(ctx context.Context, msg *servicebus.Message) error {
			log.Printf("Received message %d/%d, processing...", count, deadletterCount)
			if err := topic.Send(ctx, servicebus.NewMessageFromString(string(msg.Data))); err != nil {
				log.Println("error sending to topic", err)
				return msg.Abandon(ctx)
			}
			log.Printf("Message %d/%d sent successfully", count, deadletterCount)
			return msg.Complete(ctx)
		}))
		if err != nil {
			log.Println(err)
		}
	}
}

func ensureTopic(ctx context.Context, tm *servicebus.TopicManager, name string, opts ...servicebus.TopicManagementOption) (*servicebus.TopicEntity, error) {
	te, err := tm.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	if te == nil {
		te, err = tm.Put(ctx, name, opts...)
		if err != nil {
			return nil, err
		}
	}

	return te, nil
}

func ensureSubscription(ctx context.Context, sm *servicebus.SubscriptionManager, name string, opts ...servicebus.SubscriptionManagementOption) (*servicebus.SubscriptionEntity, error) {
	subEntity, err := sm.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	if subEntity == nil {
		subEntity, err = sm.Put(ctx, name, opts...)
		if err != nil {
			return nil, err
		}
	}

	return subEntity, nil
}

func JSONStringPretty(v interface{}) string {
	jb, err := json.MarshalIndent(v, "", " ")

	if err != nil {
		return err.Error()
	}

	return string(jb)
}
