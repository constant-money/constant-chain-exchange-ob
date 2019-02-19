package services

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
)

type PubSub struct {
	client *pubsub.Client
	prefix string
}

func InitPubSub(projectId string, prefix string) *PubSub {
	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		fmt.Println("pubsub Init err", err.Error())
		return nil
	}

	pubsub := &PubSub{
		client: pubsubClient,
		prefix: prefix,
	}

	return pubsub
}

func (ps *PubSub) GetOrCreateTopic(topicName string) *pubsub.Topic {
	prefixTopicName := topicName
	if ps.prefix != "" {
		prefixTopicName = fmt.Sprintf("%s_%s", ps.prefix, topicName)
	}

	ctx := context.Background()
	var topic *pubsub.Topic
	var err error
	fmt.Println("---------- Topic ----------")
	fmt.Println(prefixTopicName)
	fmt.Println("---------- Topic ----------")
	topic = ps.client.Topic(prefixTopicName)
	ok, err := topic.Exists(ctx)

	if err != nil {
		fmt.Println("Topic.Exists error", err.Error())
		return nil
	}

	if ok {
		return topic
	}

	fmt.Println("Creating topic", prefixTopicName)
	// create topic
	topic, err = ps.client.CreateTopic(context.Background(), prefixTopicName)
	if err != nil {
		fmt.Println("pubsub GetOrCreateTopic err", err.Error())
		return nil
	}

	return topic
}

func (ps *PubSub) GetOrCreateSubscription(subscriptionName string, topic *pubsub.Topic) *pubsub.Subscription {
	prefixSubscriptionName := subscriptionName
	if ps.prefix != "" {
		prefixSubscriptionName = fmt.Sprintf("%s_%s", ps.prefix, subscriptionName)
	}

	ctx := context.Background()
	var subscription *pubsub.Subscription
	var err error
	subscription = ps.client.Subscription(prefixSubscriptionName)

	ok, err := subscription.Exists(ctx)
	if err != nil {
		fmt.Println("Subscription.Exists error", err.Error())
		return nil
	}

	if ok {
		return subscription
	}

	fmt.Println("Creating subscription", prefixSubscriptionName)
	// create topic
	subscription, err = ps.client.CreateSubscription(ctx, prefixSubscriptionName, pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		fmt.Println("pubsub GetOrCreateSubscription err", err.Error())
		return nil
	}

	return subscription
}
