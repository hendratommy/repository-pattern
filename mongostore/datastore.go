package mongostore

import (
	"context"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"github.com/hendratommy/repository-pattern/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PostCollection    = "posts"
	CommentCollection = "comments"
)

func FindByID(ctx context.Context, coll *mongo.Collection, id int, m interface{}) error {
	res := coll.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return err
	}
	return res.Decode(m)
}

func FindPostByID(ctx context.Context, db *mongo.Database, id int) (*models.Post, error) {
	p := new(models.Post)
	err := FindByID(ctx, db.Collection(PostCollection), id, p)
	return p, err
}

func SavePost(ctx context.Context, db *mongo.Database, p *models.Post) error {
	if p.ID == 0 {
		id, err := sequence.NextVal("postSeq")
		if err != nil {
			return err
		}
		p.ID = id
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	var doc bson.M
	err := db.Collection(PostCollection).FindOneAndReplace(ctx, bson.D{{"_id", p.ID}}, p, opts).Decode(&doc)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	return nil
}

func FindCommentsByPostID(ctx context.Context, db *mongo.Database, postID int) ([]*models.Comment, error) {
	cur, err := db.Collection(CommentCollection).Find(ctx, bson.D{{ "post_id", postID }})
	if err != nil {
		return nil, err
	}
	var comments []*models.Comment
	err = cur.All(ctx, &comments)
	return comments, err
}

func SaveComment(ctx context.Context, db *mongo.Database, c *models.Comment) error {
	if c.ID == 0 {
		id, err := sequence.NextVal("commentSeq")
		if err != nil {
			return err
		}
		c.ID = id
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	var doc bson.M
	err := db.Collection(CommentCollection).FindOneAndReplace(ctx, bson.D{{"_id", c.ID}}, c, opts).Decode(&doc)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	return nil
}