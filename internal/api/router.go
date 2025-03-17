package api

import (
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/brightfame/metamorph/internal/api/handlers"
	"github.com/brightfame/metamorph/internal/api/middleware"
	"github.com/brightfame/metamorph/internal/changeset"
	"github.com/brightfame/metamorph/internal/config"
	"github.com/brightfame/metamorph/internal/executor"
)

func SetupRouter(db *gorm.DB, changesetService *changeset.Service, execService *executor.ExecutionService, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.Default())

	// Add authentication middleware if configured
	if cfg.GitHubToken != "" {
		router.Use(middleware.Auth())
	}

	// Setup static files
	router.Static("/assets", "./web/assets")

	// Setup HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Initialize handlers
	changesetHandler := handlers.NewChangesetHandler(changesetService, execService)

	// API routes
	api := router.Group("/api")
	{
		changesets := api.Group("/changesets")
		{
			changesets.GET("", changesetHandler.ListChangesets)
			changesets.POST("", changesetHandler.CreateChangeset)
			changesets.GET("/:id", changesetHandler.GetChangeset)
			changesets.PUT("/:id", changesetHandler.UpdateChangeset)
			changesets.POST("/:id/publish", changesetHandler.PublishChangeset)
			changesets.GET("/:id/preview", changesetHandler.PreviewChangeset)
		}
	}

	// Web UI routes
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/changesets")
	})

	router.GET("/changesets", func(c *gin.Context) {
		changesets, err := changesetService.GetAllChangesets()
		if err != nil {
			c.HTML(500, "error.html", gin.H{"error": err.Error()})
			return
		}

		c.HTML(200, "changesets/list.html", gin.H{
			"title":      "MetaMorph - Changesets",
			"changesets": changesets,
		})
	})

	router.GET("/changesets/new", func(c *gin.Context) {
		c.HTML(200, "changesets/new.html", gin.H{
			"title": "MetaMorph - New Changeset",
		})
	})

	router.GET("/changesets/:id", func(c *gin.Context) {
		id := c.Param("id")
		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.HTML(400, "error.html", gin.H{"error": "Invalid ID"})
			return
		}

		changeset, err := changesetService.GetChangesetByID(uint(idUint))
		if err != nil {
			c.HTML(404, "error.html", gin.H{"error": "Changeset not found"})
			return
		}

		c.HTML(200, "changesets/view.html", gin.H{
			"title":     "MetaMorph - " + changeset.Name,
			"changeset": changeset,
		})
	})

	router.GET("/changesets/:id/edit", func(c *gin.Context) {
		id := c.Param("id")
		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.HTML(400, "error.html", gin.H{"error": "Invalid ID"})
			return
		}

		changeset, err := changesetService.GetChangesetByID(uint(idUint))
		if err != nil {
			c.HTML(404, "error.html", gin.H{"error": "Changeset not found"})
			return
		}

		// Only allow editing of draft changesets
		if changeset.Status != changeset.Draft {
			c.HTML(400, "error.html", gin.H{"error": "Only draft changesets can be edited"})
			return
		}

		c.HTML(200, "changesets/edit.html", gin.H{
			"title":     "MetaMorph - Edit " + changeset.Name,
			"changeset": changeset,
		})
	})

	return router
}
