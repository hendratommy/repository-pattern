package repositories

import (
	"context"
	"errors"
	"github.com/hendratommy/repository-pattern/models"
	"github.com/hendratommy/repository-pattern/sqlstore"
	"github.com/jmoiron/sqlx"
)

type ctxTransactionKey struct{}

type SqlPostRepository struct {
	db *sqlx.DB
}

type sqlRepository interface {
	getDB() *sqlx.DB
}

var ErrInvalidTxType = errors.New("invalid tx type, tx type should be *sqlx.Tx")

func getSqlxDatabase(ctx context.Context, r sqlRepository) (sqlstore.SqlxDatabase, error) {
	txv := ctx.Value(ctxTransactionKey{})
	if txv == nil {
		return r.getDB(), nil
	}
	if tx, ok := txv.(*sqlx.Tx); ok {
		return tx, nil
	}
	return nil, ErrInvalidTxType
}

func inSqlTransaction(ctx context.Context, r sqlRepository, fn func(context.Context) error) error {
	tx, err := r.getDB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	trxCtx := context.WithValue(ctx, ctxTransactionKey{}, tx)

	err = fn(trxCtx)
	if err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

func NewSqlPostRepository(db *sqlx.DB) *SqlPostRepository {
	return &SqlPostRepository{db: db}
}

func (r *SqlPostRepository) getDB() *sqlx.DB {
	return r.db
}

func (r *SqlPostRepository) Save(ctx context.Context, p *models.Post) error {
	db, err := getSqlxDatabase(ctx, r)
	if err != nil {
		return err
	}
	return sqlstore.SavePost(ctx, db, p)
}

func (r *SqlPostRepository) FindByID(ctx context.Context, id int) (*models.Post, error) {
	db, err := getSqlxDatabase(ctx, r)
	if err != nil {
		return nil, err
	}
	return sqlstore.FindPostByID(ctx, db, id)
}

func (r *SqlPostRepository) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return inSqlTransaction(ctx, r, fn)
}

type SqlCommentRepository struct {
	db *sqlx.DB
}

func NewSqlCommentRepository(db *sqlx.DB) *SqlCommentRepository {
	return &SqlCommentRepository{db: db}
}

func (r *SqlCommentRepository) getDB() *sqlx.DB {
	return r.db
}

func (r *SqlCommentRepository) Save(ctx context.Context, c *models.Comment) error {
	db, err := getSqlxDatabase(ctx, r)
	if err != nil {
		return err
	}
	return sqlstore.SaveComment(ctx, db, c)
}

func (r *SqlCommentRepository) FindByPostID(ctx context.Context, postID int) ([]*models.Comment, error) {
	db, err := getSqlxDatabase(ctx, r)
	if err != nil {
		return nil, err
	}
	return sqlstore.FindCommentsByPostID(ctx, db, postID)
}

func (r *SqlCommentRepository) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return inSqlTransaction(ctx, r, fn)
}
