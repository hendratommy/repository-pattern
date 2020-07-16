package models

import (
	"context"
)

type PostRepository interface {
	Save(ctx context.Context, p *Post) error
	FindByID(ctx context.Context, id int) (*Post, error)
	InTransaction(ctx context.Context, fn func(context.Context) error) error
}

type CommentRepository interface {
	Save(ctx context.Context, c *Comment) error
	FindByPostID(ctx context.Context, postID int) ([]*Comment, error)
	InTransaction(ctx context.Context, fn func(context.Context) error) error
}
