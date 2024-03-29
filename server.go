package main

import (
	"errors"
	"html/template"
	"io"

	"github.com/champbronc2/buzz-me/handler"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/mgo.v2"
)

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// Define the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

func main() {
	e := echo.New()

	templates := make(map[string]*template.Template)
	templates["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
	templates["signup.html"] = template.Must(template.ParseFiles("templates/signup.html", "templates/base.html"))
	templates["user.html"] = template.Must(template.ParseFiles("templates/user.html", "templates/base.html"))
	templates["login.html"] = template.Must(template.ParseFiles("templates/login.html", "templates/base.html"))
	templates["dashboard.html"] = template.Must(template.ParseFiles("templates/dashboard.html", "templates/base.html"))
	templates["post.html"] = template.Must(template.ParseFiles("templates/post.html", "templates/base.html"))
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Logger())
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		TokenLookup: "cookie:" + echo.HeaderAuthorization,
		SigningKey:  []byte(handler.Key),
		Skipper: func(c echo.Context) bool {
			// Skip authentication for requests
			if c.Path() != "/dashboard" {
				return true
			}
			return false
		},
	}))

	// Database connection
	db, err := mgo.Dial("localhost")
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Create indices
	if err = db.Copy().DB("buzzme").C("users").EnsureIndex(mgo.Index{
		Key:    []string{"username"},
		Unique: true,
	}); err != nil {
		log.Fatal(err)
	}

	// Initialize handler
	h := &handler.Handler{DB: db}

	// Routes
	e.GET("/", h.Index)
	e.GET("/post/:id", h.FetchPost)
	e.GET("/dashboard", h.Dashboard)
	e.GET("/list", h.ListUsers)
	e.GET("/:username", h.FetchUser)
	e.GET("/signup", h.ViewSignup)
	e.GET("/login", h.ViewLogin)

	e.PUT("/dashboard", h.UpdateUser)

	e.POST("/post", h.CheckPost)
	e.POST("/signup", h.Signup)
	e.POST("/login", h.Login)
	e.POST("/posts", h.CreatePost)
	e.POST("/dashboard", h.CreateWithdrawal)

	e.Static("/static", "static")
	e.File("/favicon.ico", "static/images/favicon.ico")

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
