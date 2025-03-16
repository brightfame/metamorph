package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start the HTTP server to serve the application`,
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		fmt.Printf("Starting server on port %s...\n", port)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Welcome to the server!")
		})

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	},
}
