package repositories

import (
	"context"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type VerificationRepo struct {
	collection *mongo.Collection
}

func NewVerificationRepo(collection *mongo.Collection) *VerificationRepo {
	return &VerificationRepo{collection: collection}
}

func (r *VerificationRepo) StoreVerification(ctx context.Context, v userpkg.Verification) error {
	_, err := r.collection.InsertOne(ctx, v)
	return err
}

func (r *VerificationRepo) GetVerification(ctx context.Context, email string) (userpkg.Verification, error) {
	var v userpkg.Verification
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&v)
	return v, err
}

func (r *VerificationRepo) DeleteVerification(ctx context.Context, email string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"email": email})
	return err
}

func (r *VerificationRepo) IncrementAttemptCount(ctx context.Context, email string) error {
	   _, err := r.collection.UpdateOne(
		   ctx,
		   bson.M{"email": email},
		   // increment the attemptCount field matching struct tag
		   bson.M{"$inc": bson.M{"attemptCount": 1}},
	   )
	return err
}
