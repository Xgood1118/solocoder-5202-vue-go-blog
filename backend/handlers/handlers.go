package handlers

import (
	"blog-backend/models"
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/feeds"
	"github.com/yuin/goldmark"
)

type Handler struct {
	store *models.Store
}

func NewHandler(store *models.Store) *Handler {
	return &Handler{store: store}
}

func markdownToHTML(content string) string {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return content
	}
	return buf.String()
}

func countWords(content string) int {
	return utf8.RuneCountInString(strings.ReplaceAll(content, " ", ""))
}

func calculateReadTime(wordCount int) int {
	return (wordCount + 199) / 200
}

func (h *Handler) GetArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	tag := c.Query("tag")
	categoryID := c.Query("category")
	search := c.Query("search")
	author := c.Query("author")
	sortBy := c.DefaultQuery("sortBy", "time")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	var articles []models.Article
	h.store.Articles.Range(func(_, value interface{}) bool {
		article := value.(models.Article)

		if tag != "" && !contains(article.Tags, tag) {
			return true
		}
		if categoryID != "" && article.CategoryID != categoryID {
			return true
		}
		if author != "" && article.Author != author {
			return true
		}
		if search != "" {
			lowerSearch := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(article.Title), lowerSearch) &&
				!strings.Contains(strings.ToLower(article.Content), lowerSearch) &&
				!containsLower(article.Tags, lowerSearch) {
				return true
			}
		}

		articles = append(articles, article)
		return true
	})

	if sortBy == "views" {
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].ViewCount > articles[j].ViewCount
		})
	} else {
		sort.Slice(articles, func(i, j int) bool {
			a, b := articles[i], articles[j]
			if a.IsPinned && b.IsPinned {
				if a.PinnedAt.Equal(b.PinnedAt) {
					return a.CreatedAt.After(b.CreatedAt)
				}
				return a.PinnedAt.After(b.PinnedAt)
			}
			if a.IsPinned {
				return true
			}
			if b.IsPinned {
				return false
			}
			return a.CreatedAt.After(b.CreatedAt)
		})
	}

	total := len(articles)
	start := (page - 1) * size
	end := start + size
	if start >= total {
		c.JSON(http.StatusOK, gin.H{"list": []models.Article{}, "total": total})
		return
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{"list": articles[start:end], "total": total})
}

func (h *Handler) CreateArticle(c *gin.Context) {
	var req struct {
		Title      string   `json:"title" binding:"required"`
		Author     string   `json:"author" binding:"required"`
		CategoryID string   `json:"categoryId" binding:"required"`
		Tags       []string `json:"tags"`
		Content    string   `json:"content" binding:"required"`
		IsPinned   bool     `json:"isPinned"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": 400})
		return
	}

	if _, ok := h.store.Categories.Load(req.CategoryID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found", "code": 400})
		return
	}

	now := time.Now()
	wordCount := countWords(req.Content)
	article := models.Article{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Author:      req.Author,
		CreatedAt:   now,
		UpdatedAt:   now,
		CategoryID:  req.CategoryID,
		Tags:        req.Tags,
		Content:     req.Content,
		ContentHTML: markdownToHTML(req.Content),
		IsPinned:    req.IsPinned,
		PinnedAt:    now,
		ViewCount:   0,
		LikeCount:   0,
		CommentCount: 0,
		WordCount:   wordCount,
		ReadTime:    calculateReadTime(wordCount),
	}

	h.store.Articles.Store(article.ID, article)
	c.JSON(http.StatusCreated, article)
}

func (h *Handler) GetArticle(c *gin.Context) {
	id := c.Param("id")

	val, ok := h.store.Articles.Load(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found", "code": 404})
		return
	}

	article := val.(models.Article)
	atomic.AddInt64(&article.ViewCount, 1)
	h.store.Articles.Store(id, article)

	var comments []models.Comment
	h.store.Comments.Range(func(_, value interface{}) bool {
		comment := value.(models.Comment)
		if comment.ArticleID == id {
			comments = append(comments, comment)
		}
		return true
	})

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	var related []models.Article
	h.store.Articles.Range(func(_, value interface{}) bool {
		other := value.(models.Article)
		if other.ID != id {
			overlap := 0
			for _, t := range article.Tags {
				if contains(other.Tags, t) {
					overlap++
				}
			}
			if overlap > 0 {
				related = append(related, other)
			}
		}
		return true
	})

	sort.Slice(related, func(i, j int) bool {
		return len(related[i].Tags) > len(related[j].Tags)
	})
	if len(related) > 5 {
		related = related[:5]
	}

	c.JSON(http.StatusOK, gin.H{
		"article":  article,
		"comments": comments,
		"related":  related,
	})
}

func (h *Handler) UpdateArticle(c *gin.Context) {
	id := c.Param("id")

	val, ok := h.store.Articles.Load(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found", "code": 404})
		return
	}

	var req struct {
		Title      string   `json:"title"`
		CategoryID string   `json:"categoryId"`
		Tags       []string `json:"tags"`
		Content    string   `json:"content"`
		IsPinned   *bool    `json:"isPinned"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": 400})
		return
	}

	article := val.(models.Article)
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.CategoryID != "" {
		if _, ok := h.store.Categories.Load(req.CategoryID); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found", "code": 400})
			return
		}
		article.CategoryID = req.CategoryID
	}
	if req.Tags != nil {
		article.Tags = req.Tags
	}
	if req.Content != "" {
		article.Content = req.Content
		article.ContentHTML = markdownToHTML(req.Content)
		article.WordCount = countWords(req.Content)
		article.ReadTime = calculateReadTime(article.WordCount)
	}
	if req.IsPinned != nil {
		article.IsPinned = *req.IsPinned
		if *req.IsPinned {
			article.PinnedAt = time.Now()
		}
	}
	article.UpdatedAt = time.Now()

	h.store.Articles.Store(id, article)
	c.JSON(http.StatusOK, article)
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	if _, ok := h.store.Articles.Load(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found", "code": 404})
		return
	}

	h.store.Articles.Delete(id)

	var toDelete []string
	h.store.Comments.Range(func(key, value interface{}) bool {
		if value.(models.Comment).ArticleID == id {
			toDelete = append(toDelete, key.(string))
		}
		return true
	})
	for _, k := range toDelete {
		h.store.Comments.Delete(k)
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateComment(c *gin.Context) {
	articleID := c.Param("id")

	if _, ok := h.store.Articles.Load(articleID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found", "code": 404})
		return
	}

	var req struct {
		Author   string `json:"author" binding:"required"`
		Content  string `json:"content" binding:"required"`
		ParentID string `json:"parentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": 400})
		return
	}

	comment := models.Comment{
		ID:         uuid.New().String(),
		ArticleID:  articleID,
		Author:     req.Author,
		Content:    req.Content,
		CreatedAt:  time.Now(),
		ParentID:   req.ParentID,
		LikeCount:  0,
		IsReported: false,
	}

	h.store.Comments.Store(comment.ID, comment)

	if val, ok := h.store.Articles.Load(articleID); ok {
		article := val.(models.Article)
		atomic.AddInt64(&article.CommentCount, 1)
		h.store.Articles.Store(articleID, article)
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) LikeArticle(c *gin.Context) {
	id := c.Param("id")
	clientID := getClientID(c)

	val, ok := h.store.Articles.Load(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found", "code": 404})
		return
	}

	likeKey := fmt.Sprintf("article:%s:%s", id, clientID)
	if _, liked := h.store.Likes.LoadOrStore(likeKey, true); liked {
		c.JSON(http.StatusConflict, gin.H{"error": "Already liked", "code": 409})
		return
	}

	article := val.(models.Article)
	atomic.AddInt64(&article.LikeCount, 1)
	h.store.Articles.Store(id, article)

	c.JSON(http.StatusOK, gin.H{"likeCount": article.LikeCount})
}

func (h *Handler) GetCategories(c *gin.Context) {
	var categories []models.Category
	h.store.Categories.Range(func(_, value interface{}) bool {
		categories = append(categories, value.(models.Category))
		return true
	})
	c.JSON(http.StatusOK, categories)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": 400})
		return
	}

	var exists bool
	h.store.Categories.Range(func(_, value interface{}) bool {
		cat := value.(models.Category)
		if cat.Slug == req.Slug || cat.Name == req.Name {
			exists = true
			return false
		}
		return true
	})

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Category already exists", "code": 409})
		return
	}

	category := models.Category{
		ID:   uuid.New().String(),
		Name: req.Name,
		Slug: req.Slug,
	}

	h.store.Categories.Store(category.ID, category)
	c.JSON(http.StatusCreated, category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	val, ok := h.store.Categories.Load(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found", "code": 404})
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "code": 400})
		return
	}

	category := val.(models.Category)
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Slug != "" {
		category.Slug = req.Slug
	}

	h.store.Categories.Store(id, category)
	c.JSON(http.StatusOK, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	if _, ok := h.store.Categories.Load(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found", "code": 404})
		return
	}

	var hasArticles bool
	h.store.Articles.Range(func(_, value interface{}) bool {
		if value.(models.Article).CategoryID == id {
			hasArticles = true
			return false
		}
		return true
	})

	if hasArticles {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category with articles", "code": 409})
		return
	}

	h.store.Categories.Delete(id)
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetTagCloud(c *gin.Context) {
	tagCount := make(map[string]int)
	h.store.Articles.Range(func(_, value interface{}) bool {
		for _, tag := range value.(models.Article).Tags {
			tagCount[tag]++
		}
		return true
	})

	type TagInfo struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var tags []TagInfo
	for name, count := range tagCount {
		tags = append(tags, TagInfo{Name: name, Count: count})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Count > tags[j].Count
	})

	if len(tags) > 30 {
		tags = tags[:30]
	}

	c.JSON(http.StatusOK, tags)
}

func (h *Handler) GetActiveAuthors(c *gin.Context) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	type authorStat struct {
		Posts     int
		ViewCount int64
	}
	authorStats := make(map[string]authorStat)

	h.store.Articles.Range(func(_, value interface{}) bool {
		article := value.(models.Article)
		if article.CreatedAt.After(thirtyDaysAgo) {
			stat := authorStats[article.Author]
			stat.Posts++
			stat.ViewCount += article.ViewCount
			authorStats[article.Author] = stat
		}
		return true
	})

	type AuthorInfo struct {
		Name  string  `json:"name"`
		Score float64 `json:"score"`
		Posts int     `json:"posts"`
		Views int64   `json:"views"`
	}

	var authors []AuthorInfo
	for name, stats := range authorStats {
		score := float64(stats.Posts)*100 + float64(stats.ViewCount)/10
		authors = append(authors, AuthorInfo{
			Name:  name,
			Score: score,
			Posts: stats.Posts,
			Views: stats.ViewCount,
		})
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Score > authors[j].Score
	})

	if len(authors) > 10 {
		authors = authors[:10]
	}

	c.JSON(http.StatusOK, authors)
}

func (h *Handler) GetArchive(c *gin.Context) {
	type MonthArchive struct {
		Month    string           `json:"month"`
		Articles []models.Article `json:"articles"`
	}

	archiveMap := make(map[string][]models.Article)
	h.store.Articles.Range(func(_, value interface{}) bool {
		article := value.(models.Article)
		month := article.CreatedAt.Format("2006-01")
		archiveMap[month] = append(archiveMap[month], article)
		return true
	})

	var archives []MonthArchive
	for month, articles := range archiveMap {
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].CreatedAt.After(articles[j].CreatedAt)
		})
		archives = append(archives, MonthArchive{Month: month, Articles: articles})
	}

	sort.Slice(archives, func(i, j int) bool {
		return archives[i].Month > archives[j].Month
	})

	c.JSON(http.StatusOK, archives)
}

func (h *Handler) LikeComment(c *gin.Context) {
	id := c.Param("id")
	clientID := getClientID(c)

	val, ok := h.store.Comments.Load(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found", "code": 404})
		return
	}

	likeKey := fmt.Sprintf("comment:%s:%s", id, clientID)
	if _, liked := h.store.Likes.LoadOrStore(likeKey, true); liked {
		c.JSON(http.StatusConflict, gin.H{"error": "Already liked", "code": 409})
		return
	}

	comment := val.(models.Comment)
	atomic.AddInt64(&comment.LikeCount, 1)
	h.store.Comments.Store(id, comment)

	c.JSON(http.StatusOK, gin.H{"likeCount": comment.LikeCount})
}

func (h *Handler) ReportComment(c *gin.Context) {
	id := c.Param("id")

	if _, ok := h.store.Comments.Load(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found", "code": 404})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	report := models.Report{
		ID:        uuid.New().String(),
		CommentID: id,
		Reason:    req.Reason,
		CreatedAt: time.Now(),
		Handled:   false,
	}

	h.store.Reports.Store(report.ID, report)

	if val, ok := h.store.Comments.Load(id); ok {
		comment := val.(models.Comment)
		comment.IsReported = true
		h.store.Comments.Store(id, comment)
	}

	c.JSON(http.StatusCreated, report)
}

func (h *Handler) GetReports(c *gin.Context) {
	var reports []models.Report
	h.store.Reports.Range(func(_, value interface{}) bool {
		reports = append(reports, value.(models.Report))
		return true
	})

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].CreatedAt.After(reports[j].CreatedAt)
	})

	c.JSON(http.StatusOK, reports)
}

func (h *Handler) GetRSS(c *gin.Context) {
	feed := &feeds.Feed{
		Title:       "Internal Tech Blog",
		Link:        &feeds.Link{Href: "http://localhost:5173"},
		Description: "Technical share and notes from the team",
		Created:     time.Now(),
	}

	var articles []models.Article
	h.store.Articles.Range(func(_, value interface{}) bool {
		articles = append(articles, value.(models.Article))
		return true
	})

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].CreatedAt.After(articles[j].CreatedAt)
	})

	if len(articles) > 20 {
		articles = articles[:20]
	}

	for _, a := range articles {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       a.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("http://localhost:5173/article/%s", a.ID)},
			Description: a.ContentHTML,
			Author:      &feeds.Author{Name: a.Author},
			Created:     a.CreatedAt,
		})
	}

	rss, err := feed.ToRss()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rss)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsLower(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), item) {
			return true
		}
	}
	return false
}

func getClientID(c *gin.Context) string {
	cookie, err := c.Cookie("client_id")
	if err != nil {
		cookie = uuid.New().String()
		c.SetCookie("client_id", cookie, 365*24*60*60, "/", "", false, true)
	}
	return cookie
}
