package repository_pattern

import (
	"context"
	"errors"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"github.com/hendratommy/repository-pattern/models"
	"github.com/hendratommy/repository-pattern/mongostore"
	"github.com/hendratommy/repository-pattern/repositories"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

func try(err error) {
	if err != nil {
		panic(err)
	}
}

func prepareMongoEnvironment(db *mongo.Database) {
	try(db.Drop(context.Background()))
	try(db.CreateCollection(context.Background(), mongostore.PostCollection))
	try(db.CreateCollection(context.Background(), mongostore.CommentCollection))
	sequence.SetupDefaultSequence(db, 30*time.Second)
}

func TestMongoRepository(t *testing.T) {
	var dbName = "repositoryPattern"
	var client, err = mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	if err = client.Connect(context.TODO()); err != nil {
		panic(err)
	}

	var db = client.Database(dbName)
	//defer db.Drop(context.Background()) // clean up test database

	Convey("Test mongo repository", t, func() {
		prepareMongoEnvironment(db)

		postRepo := repositories.NewMongoPostRepository(db)
		commentRepo := repositories.NewMongoCommentRepository(db)

		Convey("Test transaction commit", func() {
			var p *models.Post
			err := postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
				p = &models.Post{
					Title: "implement repository pattern in go",
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
				return commentRepo.Save(ctx, c2)
			})

			Convey("Should commit successfully", func() {
				So(err, ShouldBeNil)

				var posts []bson.M
				cur, err := db.Collection(mongostore.PostCollection).Find(context.Background(), bson.D{})
				try(err)
				try(cur.All(context.Background(), &posts))

				So(len(posts), ShouldEqual, 1)
				So(posts[0]["_id"], ShouldEqual, p.ID)

				var comments []bson.M
				cur, err = db.Collection(mongostore.CommentCollection).Find(context.Background(), bson.D{})
				try(err)
				try(cur.All(context.Background(), &comments))

				So(len(comments), ShouldEqual, 2)
				So(comments[0]["post_id"], ShouldEqual, p.ID)
				So(comments[1]["post_id"], ShouldEqual, p.ID)
			})
		})

		Convey("Test transaction rollback", func() {
			err := postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
				p := &models.Post{
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

				return errors.New("should rollback")
			})

			Convey("Should rollback successfully", func() {
				So(err, ShouldBeNil)

				var posts []bson.M
				cur, err := db.Collection(mongostore.PostCollection).Find(context.Background(), bson.D{})
				try(err)
				try(cur.All(context.Background(), &posts))

				So(len(posts), ShouldEqual, 0)

				var comments []bson.M
				cur, err = db.Collection(mongostore.CommentCollection).Find(context.Background(), bson.D{})
				try(err)
				try(cur.All(context.Background(), &comments))

				So(len(comments), ShouldEqual, 0)
			})
		})

		Convey("Test nested transactions", func() {

			Convey("Test nested transaction commit", func() {
				err := postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
					p := &models.Post{
						Title: "implement repository pattern in go",
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

					return postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
						p := &models.Post{
							Title: "implement repository pattern in go",
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
					})
				})

				Convey("Should commit successfully", func() {
					So(err, ShouldBeNil)

					var posts []bson.M
					cur, err := db.Collection(mongostore.PostCollection).Find(context.Background(), bson.D{})
					try(err)
					try(cur.All(context.Background(), &posts))

					So(len(posts), ShouldEqual, 2)

					var comments []bson.M
					cur, err = db.Collection(mongostore.CommentCollection).Find(context.Background(), bson.D{})
					try(err)
					try(cur.All(context.Background(), &comments))

					So(len(comments), ShouldEqual, 4)
				})
			})

			Convey("Test nested transaction partial commit", func() {
				err := postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
					p := &models.Post{
						Title: "implement repository pattern in go",
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

					return postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
						p := &models.Post{
							Title: "implement repository pattern in go",
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
						return errors.New("should rollback")
					})
				})

				Convey("Should partial commit successfully", func() {
					So(err, ShouldBeNil)

					var posts []bson.M
					cur, err := db.Collection(mongostore.PostCollection).Find(context.Background(), bson.D{})
					try(err)
					try(cur.All(context.Background(), &posts))

					So(len(posts), ShouldEqual, 1)

					var comments []bson.M
					cur, err = db.Collection(mongostore.CommentCollection).Find(context.Background(), bson.D{})
					try(err)
					try(cur.All(context.Background(), &comments))

					So(len(comments), ShouldEqual, 2)
				})
			})

			Convey("Test nested transaction rollback", func() {
				err := postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
					var err error

					p := &models.Post{
						Title: "implement repository pattern in go",
					}
					if err = postRepo.Save(ctx, p); err != nil {
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
					if err = commentRepo.Save(ctx, c1); err != nil {
						return err
					}
					if err = commentRepo.Save(ctx, c2); err != nil {
						return err
					}

					postRepo.InTransaction(context.Background(), func(ctx context.Context) error {
						p := &models.Post{
							Title: "implement repository pattern in go",
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
						if err = commentRepo.Save(ctx, c1); err != nil {
							return err
						}
						if err = commentRepo.Save(ctx, c2); err != nil {
							return err
						}
						err = errors.New("should rollback")
						return err
					})

					return err
				})

				Convey("Should rollback successfully", func() {
					So(err, ShouldBeNil)

					var posts []bson.M
					cur, err := db.Collection(mongostore.PostCollection).Find(context.Background(), bson.D{})
					try(err)
					try(cur.All(context.Background(), &posts))

					So(len(posts), ShouldEqual, 0)

					var comments []bson.M
					cur, err = db.Collection(mongostore.CommentCollection).Find(context.Background(), bson.D{})
					try(err)
					try(cur.All(context.Background(), &comments))

					So(len(comments), ShouldEqual, 0)
				})
			})

		})
	})
}
