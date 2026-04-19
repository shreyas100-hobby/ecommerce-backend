package database

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func NewFirestoreClient(ctx context.Context, credentialsFile string, credentialsJSON string) (*firestore.Client, error) {
	var opt option.ClientOption
	if credentialsJSON != "" {
		opt = option.WithCredentialsJSON([]byte(credentialsJSON))
	} else {
		opt = option.WithCredentialsFile(credentialsFile)
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %v", err)
	}

	return client, nil
}
