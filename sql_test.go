package repository_pattern

import (
	"context"
	"errors"
	"github.com/hendratommy/repository-pattern/models"
	"github.com/hendratommy/repository-pattern/repositories"
	"github.com/hendratommy/repository-pattern/sqlstore"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func prepareSqlEnvironment(db *sqlx.DB) {
	sqlstore.DropTables(db)
	sqlstore.CreateTables(db)
}

func TestSqlRepository(t *testing.T) {
	db, err := sqlx.Connect("postgres", os.Getenv("PG_URI"))
	if err != nil {
		panic(err)
	}

	//defer DropTables(db) // clean up test database

	Convey("Test sql repository", t, func() {
		prepareSqlEnvironment(db)

		postRepo := repositories.NewSqlPostRepository(db)
		commentRepo := repositories.NewSqlCommentRepository(db)

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

				var posts []*models.Post
				try(db.Select(&posts, `SELECT * FROM ` + sqlstore.PostTable))

				So(len(posts), ShouldEqual, 1)
				So(posts[0].ID, ShouldEqual, p.ID)

				var comments []*models.Comment
				try(db.Select(&comments, `SELECT * FROM ` + sqlstore.CommentTable))

				So(len(comments), ShouldEqual, 2)
				So(comments[0].PostID, ShouldEqual, p.ID)
				So(comments[1].PostID, ShouldEqual, p.ID)
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

				var posts []*models.Post
				try(db.Select(&posts, `SELECT * FROM ` + sqlstore.PostTable))

				So(len(posts), ShouldEqual, 0)

				var comments []*models.Comment
				try(db.Select(&comments, `SELECT * FROM ` + sqlstore.CommentTable))

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

					var posts []*models.Post
					try(db.Select(&posts, `SELECT * FROM ` + sqlstore.PostTable))

					So(len(posts), ShouldEqual, 2)

					var comments []*models.Comment
					try(db.Select(&comments, `SELECT * FROM ` + sqlstore.CommentTable))

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

					var posts []*models.Post
					try(db.Select(&posts, `SELECT * FROM ` + sqlstore.PostTable))

					So(len(posts), ShouldEqual, 1)

					var comments []*models.Comment
					try(db.Select(&comments, `SELECT * FROM ` + sqlstore.CommentTable))

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

					var posts []*models.Post
					try(db.Select(&posts, `SELECT * FROM ` + sqlstore.PostTable))

					So(len(posts), ShouldEqual, 0)

					var comments []*models.Comment
					try(db.Select(&comments, `SELECT * FROM ` + sqlstore.CommentTable))

					So(len(comments), ShouldEqual, 0)
				})
			})

		})
	})
}
