// Facade to fulfil Repository interface, all queries and others database operation/queries should go to datastore package

package repositories

import (
	"context"
	"github.com/hendratommy/repository-pattern/models"
	"github.com/hendratommy/repository-pattern/mongostore"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoPostRepository struct {
	db *mongo.Database
}

func inMongoTransaction(ctx context.Context, db *mongo.Database, fn func(context.Context) error) error {
	sess, err := db.Client().StartSession()
	if err != nil {
		return err
	}

	return mongo.WithSession(ctx, sess, func(sc mongo.SessionContext) error {
		defer sess.EndSession(context.Background())

		if err := sc.StartTransaction(); err != nil {
			return err
		}
		if err := fn(sc); err != nil {
			return sc.AbortTransaction(sc)
		}
		return sc.CommitTransaction(sc)
	})
}

func NewMongoPostRepository(db *mongo.Database) *MongoPostRepository {
	return &MongoPostRepository{db: db}
}

func (r *MongoPostRepository) FindByID(ctx context.Context, id int) (*models.Post, error) {
	return mongostore.FindPostByID(ctx, r.db, id)
}

func (r *MongoPostRepository) Save(ctx context.Context, m *models.Post) error {
	return mongostore.SavePost(ctx, r.db, m)
}

func (r *MongoPostRepository) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return inMongoTransaction(ctx, r.db, fn)
}

type MongoCommentRepository struct {
	db *mongo.Database
}

func NewMongoCommentRepository(db *mongo.Database) *MongoCommentRepository {
	return &MongoCommentRepository{db: db}
}

func (r *MongoCommentRepository) FindByPostID(ctx context.Context, postId int) ([]*models.Comment, error) {
	return mongostore.FindCommentsByPostID(ctx, r.db, postId)
}

func (r *MongoCommentRepository) Save(ctx context.Context, m *models.Comment) error {
	return mongostore.SaveComment(ctx, r.db, m)
}

func (r *MongoCommentRepository) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return inMongoTransaction(ctx, r.db, fn)
}