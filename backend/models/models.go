package models

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CategoryID  string    `json:"categoryId"`
	Tags        []string  `json:"tags"`
	Content     string    `json:"content"`
	ContentHTML string    `json:"contentHTML"`
	IsPinned    bool      `json:"isPinned"`
	PinnedAt    time.Time `json:"pinnedAt"`
	ViewCount   int64     `json:"viewCount"`
	LikeCount   int64     `json:"likeCount"`
	CommentCount int64    `json:"commentCount"`
	WordCount   int       `json:"wordCount"`
	ReadTime    int       `json:"readTime"`
}

type Comment struct {
	ID         string    `json:"id"`
	ArticleID  string    `json:"articleId"`
	Author     string    `json:"author"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	ParentID   string    `json:"parentId"`
	LikeCount  int64     `json:"likeCount"`
	IsReported bool      `json:"isReported"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Report struct {
	ID        string    `json:"id"`
	CommentID string    `json:"commentId"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
	Handled   bool      `json:"handled"`
}

type Store struct {
	Articles   *sync.Map
	Categories *sync.Map
	Comments   *sync.Map
	Reports    *sync.Map
	Likes      *sync.Map
}

func NewStore() *Store {
	return &Store{
		Articles:   &sync.Map{},
		Categories: &sync.Map{},
		Comments:   &sync.Map{},
		Reports:    &sync.Map{},
		Likes:      &sync.Map{},
	}
}

func (s *Store) InitDefaultCategories() {
	defaults := []Category{
		{ID: uuid.New().String(), Name: "前端", Slug: "frontend"},
		{ID: uuid.New().String(), Name: "后端", Slug: "backend"},
		{ID: uuid.New().String(), Name: "运维", Slug: "devops"},
		{ID: uuid.New().String(), Name: "杂项", Slug: "misc"},
	}
	for _, c := range defaults {
		s.Categories.Store(c.ID, c)
	}
}

func (s *Store) IncrementView(articleID string) {
	if val, ok := s.Articles.Load(articleID); ok {
		article := val.(Article)
		atomic.AddInt64(&article.ViewCount, 1)
		s.Articles.Store(articleID, article)
	}
}
