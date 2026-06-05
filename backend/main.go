package main

import (
	"blog-backend/config"
	"blog-backend/handlers"
	"blog-backend/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	store := models.NewStore()
	store.InitDefaultCategories()

	h := handlers.NewHandler(store)

	api := r.Group("/api")
	{
		articles := api.Group("/articles")
		{
			articles.GET("", h.GetArticles)
			articles.POST("", h.CreateArticle)
			articles.GET("/:id", h.GetArticle)
			articles.PUT("/:id", h.UpdateArticle)
			articles.DELETE("/:id", h.DeleteArticle)
			articles.POST("/:id/comments", h.CreateComment)
			articles.POST("/:id/like", h.LikeArticle)
		}

		categories := api.Group("/categories")
		{
			categories.GET("", h.GetCategories)
			categories.POST("", h.CreateCategory)
			categories.PATCH("/:id", h.UpdateCategory)
			categories.DELETE("/:id", h.DeleteCategory)
		}

		api.GET("/tags", h.GetTagCloud)
		api.GET("/authors/active", h.GetActiveAuthors)
		api.GET("/archive", h.GetArchive)
		api.GET("/reports", h.GetReports)
		api.POST("/comments/:id/like", h.LikeComment)
		api.POST("/comments/:id/report", h.ReportComment)
	}

	r.GET("/rss.xml", h.GetRSS)

	r.Run(config.Port)
}
