package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title       string    `gorm:"not null" form:"title"`
	Slug        string    `gorm:"uniqueIndex;not null" form:"slug"`
	Content     string    `gorm:"not null" form:"content"`
	PublishedAt time.Time `gorm:"not null"`
	Visible     bool      `gorm:"not null" form:"visible"`
	Excerpt     *string   `gorm:"type:text" form:"excerpt"`
	Tags        []Tag     `gorm:"many2many:post_tags;" form:"tags"`
	AuthorID    uint      `gorm:"not null" form:"authorId"`
	Author      *User     `gorm:"foreignKey:AuthorID" form:"author"`
}

// IsScheduled returns true if the post is scheduled for future publication
func (p *Post) IsScheduled(timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	return p.Visible && p.PublishedAt.After(now)
}
