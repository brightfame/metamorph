package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start the HTTP server to serve the application`,
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		fmt.Printf("Starting server on port %s...\n", port)

		// Create a default gin router
		r := gin.Default()

		// Define routes
		r.GET("/", func(c *gin.Context) {
			c.String(200, "Welcome to the server!")
		})

		// Start the server
		if err := r.Run(":" + port); err != nil {
			log.Fatal(err)
		}
	},
}
