package main

import (
	"context"
	"errors"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"github.com/hendratommy/repository-pattern/models"
	"github.com/hendratommy/repository-pattern/mongostore"
	"github.com/hendratommy/repository-pattern/repositories"
	"github.com/hendratommy/repository-pattern/sqlstore"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var postRepo models.PostRepository
	var commentRepo models.CommentRepository

	// start use mongodb
	var dbName = "repositoryPattern"
	var client, err = mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	if err = client.Connect(context.TODO()); err != nil {
		panic(err)
	}
	var mdb = client.Database(dbName)
	// setup collections
	mdb.CreateCollection(context.Background(), mongostore.PostCollection)
	mdb.CreateCollection(context.Background(), mongostore.CommentCollection)
	// setup default sequence
	sequence.SetupDefaultSequence(mdb, 30*time.Second)
	postRepo = repositories.NewMongoPostRepository(mdb)
	commentRepo = repositories.NewMongoCommentRepository(mdb)
	// end use mongodb

	// start use postgres
	sdb, err := sqlx.Connect("postgres", os.Getenv("PG_URI"))
	if err != nil {
		panic(err)
	}
	//sqlstore.DropTables(sdb)
	sqlstore.CreateTables(sdb)
	postRepo = repositories.NewSqlPostRepository(sdb)
	commentRepo = repositories.NewSqlCommentRepository(sdb)
	// end use postgres

	var p *models.Post
	// commit
	try(postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
		p = &models.Post{
			Title: "this should persisted",
		}
		if err := postRepo.Save(ctx, p); err != nil {
			return err
		}

		c1 := &models.Comment{
			PostID: p.ID,
			Review: "yayy",
		}
		c2 := &models.Comment{
			PostID: p.ID,
			Review: "nayy",
		}
		if err := commentRepo.Save(ctx, c1); err != nil {
			return err
		}
		if err := commentRepo.Save(ctx, c2); err != nil {
			return err
		}

		return err
	}))

	// rollback
	try(postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
		p = &models.Post{
			Title: "this should not persisted",
		}
		if err := postRepo.Save(ctx, p); err != nil {
			return err
		}

		c1 := &models.Comment{
			PostID: p.ID,
			Review: "yayy",
		}
		c2 := &models.Comment{
			PostID: p.ID,
			Review: "nayy",
		}
		if err := commentRepo.Save(ctx, c1); err != nil {
			return err
		}
		if err := commentRepo.Save(ctx, c2); err != nil {
			return err
		}

		return errors.New("rollback")
	}))
}
