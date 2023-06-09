package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
)

func main() {
	r := gin.New()

	tmplName := "user"
	tmplStr := "user {{ .name }} (id {{ .id }})\n"
	tmpl := template.Must(template.New(tmplName).Parse(tmplStr))
	r.SetHTMLTemplate(tmpl)
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		name := getUser(c, id)

		c.HTML(http.StatusOK, tmplName, gin.H{
			"name": name,
			"id":   id,
		})
	})
	_ = r.Run(":8080")
}

func getUser(c *gin.Context, id string) string {
	if id == "123" {
		return "otelgin tester"
	}
	return "unknown"
}
