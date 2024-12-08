package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"codeinstyle.io/captain/config"
	"codeinstyle.io/captain/db"
	"codeinstyle.io/captain/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultTimezone    = "UTC"
	DefaultChromaStyle = "paraiso-dark"
	DefaultPostPerPage = 10
	DefaultTheme       = "light"
)

type AdminHandlers struct {
	db      *gorm.DB
	config  *config.Config
	storage storage.Provider
}

func NewAdminHandlers(database *gorm.DB, cfg *config.Config) *AdminHandlers {
	var provider storage.Provider
	var err error

	switch cfg.Storage.Provider {
	case "s3":
		provider, err = storage.NewS3Provider(cfg.Storage.S3.Bucket, cfg.Storage.S3.Region, cfg.Storage.S3.Endpoint, cfg.Storage.S3.AccessKey, cfg.Storage.S3.SecretKey)
	default: // "local"
		provider, err = storage.NewLocalProvider(cfg.Storage.LocalPath)
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize storage provider: %v", err))
	}

	return &AdminHandlers{
		db:      database,
		config:  cfg,
		storage: provider,
	}
}

func (h *AdminHandlers) Index(c *gin.Context) {
	var postCount, tagCount, userCount int64
	var recentPosts []db.Post

	// Get counts
	h.db.Model(&db.Post{}).Count(&postCount)
	h.db.Model(&db.Tag{}).Count(&tagCount)
	h.db.Model(&db.User{}).Count(&userCount)

	// Get 5 most recent posts
	h.db.Order("published_at desc").Limit(5).Find(&recentPosts)

	data := gin.H{
		"title":       "Dashboard",
		"postCount":   postCount,
		"tagCount":    tagCount,
		"userCount":   userCount,
		"recentPosts": recentPosts,
	}

	data = h.addCommonData(c, data)

	c.HTML(http.StatusOK, "admin_index.tmpl", data)
}

func (h *AdminHandlers) ShowSettings(c *gin.Context) {
	settings, err := db.GetSettings(h.db)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.tmpl", h.addCommonData(c, gin.H{}))
		return
	}

	data := gin.H{
		"title":        "Site Settings",
		"settings":     settings,
		"timezones":    h.config.GetTimezones(),
		"chromaStyles": h.config.GetChromaStyles(),
	}

	data = h.addCommonData(c, data)
	c.HTML(http.StatusOK, "admin_settings.tmpl", data)
}

func (h *AdminHandlers) UpdateSettings(c *gin.Context) {
	var form db.Settings
	var errors []string

	// Get form values
	form.Title = c.PostForm("title")
	form.Subtitle = c.PostForm("subtitle")
	form.Timezone = c.PostForm("timezone")
	form.ChromaStyle = c.PostForm("chroma_style")
	form.Theme = c.PostForm("theme")
	postsPerPage := c.PostForm("posts_per_page")

	// Validate required fields
	if form.Title == "" {
		errors = append(errors, "Title is required")
	}
	if form.Subtitle == "" {
		errors = append(errors, "Subtitle is required")
	}

	// Validate timezone
	if form.Timezone != "" {
		valid := false
		for _, tz := range h.config.GetTimezones() {
			if tz == form.Timezone {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, "Invalid timezone selected")
		}
	}

	// Validate chroma style
	if form.ChromaStyle != "" {
		valid := false
		for _, style := range h.config.GetChromaStyles() {
			if style == form.ChromaStyle {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, "Invalid syntax highlighting theme selected")
		}
	}

	// Validate theme
	if form.Theme != "" && form.Theme != "light" && form.Theme != "dark" {
		errors = append(errors, "Invalid theme selected")
	}

	// Parse and validate posts per page
	if postsPerPage != "" {
		if pp, err := strconv.Atoi(postsPerPage); err != nil {
			errors = append(errors, "Posts per page must be a number")
		} else if pp < 1 || pp > 50 {
			errors = append(errors, "Posts per page must be between 1 and 50")
		} else {
			form.PostsPerPage = pp
		}
	}

	if len(errors) > 0 {
		data := gin.H{
			"settings":     form,
			"timezones":    h.config.GetTimezones(),
			"chromaStyles": h.config.GetChromaStyles(),
			"theme":        form.Theme,
			"postsPerPage": form.PostsPerPage,
			"errors":       errors,
		}
		c.HTML(http.StatusBadRequest, "admin_settings.tmpl", h.addCommonData(c, data))
		return
	}

	// Set defaults for optional fields if not provided
	if form.Timezone == "" {
		form.Timezone = DefaultTimezone
	}
	if form.ChromaStyle == "" {
		form.ChromaStyle = DefaultChromaStyle
	}
	if form.Theme == "" {
		form.Theme = DefaultTheme
	}
	if form.PostsPerPage == 0 {
		form.PostsPerPage = DefaultPostPerPage
	}

	if err := db.UpdateSettings(h.db, &form); err != nil {
		errors = append(errors, "Failed to update settings")
		data := gin.H{
			"settings":     form,
			"timezones":    h.config.GetTimezones(),
			"chromaStyles": h.config.GetChromaStyles(),
			"theme":        form.Theme,
			"postsPerPage": form.PostsPerPage,
			"errors":       errors,
		}
		c.HTML(http.StatusInternalServerError, "admin_settings.tmpl", h.addCommonData(c, data))
		return
	}

	c.Redirect(http.StatusFound, "/admin/settings")
}
