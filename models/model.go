// Domain models package, includes Repository interface
package models

type Post struct {
	ID int `db:"id" bson:"_id"`
	Title string `db:"title" bson:"title"`
}

type Comment struct {
	ID int `db:"id" bson:"_id"`
	Review string `db:"review" bson:"review"`
	PostID int `db:"post_id" bson:"post_id"`
}
