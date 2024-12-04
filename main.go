package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"codeinstyle.io/captain/cmd"
	"codeinstyle.io/captain/config"
	"codeinstyle.io/captain/db"
	"codeinstyle.io/captain/server"
	"codeinstyle.io/captain/system"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

//go:embed embedded/admin/static/css/*
//go:embed embedded/admin/static/js/*
//go:embed embedded/admin/static/fonts/*
//go:embed embedded/admin/templates/*
//go:embed embedded/public/templates/errors/*
//go:embed embedded/public/templates/*
//go:embed embedded/themes/default/static/css/*
//go:embed embedded/themes/default/static/js/*
//go:embed embedded/themes/default/templates/*
var embeddedFS embed.FS

var (
	initDevDB  bool
	configFile string
)

func main() {
	var rootCmd = &cobra.Command{Use: "captain"}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Captain v%s\n", system.Version)
			fmt.Printf("Commit: %s\n", system.Commit)
			fmt.Printf("Built: %s\n", system.Date)
		},
	}

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Runs the server",
		Run:   runServer,
	}

	runCmd.Flags().BoolVarP(&initDevDB, "init-dev-db", "i", false, "Initialize the development database with test data")

	var userCmd = &cobra.Command{
		Use:   "user",
		Short: "User management commands",
	}

	var userCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Run:   cmd.CreateUser,
	}

	var userUpdatePasswordCmd = &cobra.Command{
		Use:   "update-password",
		Short: "Update user password",
		Run:   cmd.UpdateUserPassword,
	}

	userCmd.AddCommand(userCreateCmd, userUpdatePasswordCmd)
	rootCmd.AddCommand(runCmd, userCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Load config
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Set Gin mode based on debug flag
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	database := db.InitDB(cfg)

	// Initialize development database if requested
	if initDevDB {
		if err := db.InsertTestData(database); err != nil {
			log.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Create and start server
	srv := server.New(database, cfg, embeddedFS)

	// Run the server
	if err := srv.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
