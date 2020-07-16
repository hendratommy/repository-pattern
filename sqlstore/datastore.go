package sqlstore

import (
	"context"
	"database/sql"
	"github.com/hendratommy/repository-pattern/models"
	"github.com/jmoiron/sqlx"
)

const (
	PostTable    = "posts"
	CommentTable = "comments"
)

func DropTables(db *sqlx.DB) {
	db.Exec(`DROP TABLE ` + CommentTable)
	db.Exec(`DROP TABLE ` + PostTable)

	//_, err := db.Exec(`DROP TABLE ` + sqlstore.CommentTable)
	//try(err)
	//_, err = db.Exec(`DROP TABLE ` + sqlstore.PostTable)
	//try(err)
}

func CreateTables(db *sqlx.DB) {
	db.Exec(`CREATE TABLE ` + PostTable + `(
		id serial not null primary key,
    	title varchar(250) not null
	)`)
	db.Exec(`CREATE TABLE ` + CommentTable + `(
		id serial not null primary key,
		post_id integer not null references posts(id),
		review varchar(250) not null
	)`)
}

type SqlxDatabase interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func FindPostByID(ctx context.Context, db SqlxDatabase, id int) (*models.Post, error) {
	p := new(models.Post)
	sql := `SELECT * FROM ` + PostTable + ` WHERE id=$1`
	err := db.GetContext(ctx, p, sql, id)

	return p, err
}

func SavePost(ctx context.Context, db SqlxDatabase, p *models.Post) error {
	sql := `INSERT INTO ` + PostTable + `(title) VALUES ($1) ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title RETURNING id`
	var lastId int
	stmt, err := db.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	stmt.GetContext(ctx, &lastId, p.Title)
	p.ID = lastId
	return err
}

func FindCommentsByPostID(ctx context.Context, db SqlxDatabase, postID int) ([]*models.Comment, error) {
	var comments []*models.Comment
	sql := `SELECT * FROM ` + CommentTable + ` WHERE post_id=$1`
	err := db.SelectContext(ctx, &comments, sql, postID)
	return comments, err
}

func SaveComment(ctx context.Context, db SqlxDatabase, c *models.Comment) error {
	sql := `INSERT INTO ` + CommentTable + `(review, post_id) VALUES($1, $2) ON CONFLICT(id)
			DO UPDATE SET review=EXCLUDED.review, post_id=EXCLUDED.post_id
			RETURNING id`
	var lastId int
	stmt, err := db.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	stmt.GetContext(ctx, &lastId, c.Review, c.PostID)
	c.ID = lastId
	return err
}
