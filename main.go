package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
)

type Config struct {
	TrustedProxies []string `envconfig:"TRUSTED_PROXIES"`
	Port           string   `envconfig:"PORT" default:"8080"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("failed to process config: %s", err.Error())
	}

	router, err := setupRouter(cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = router.Run(fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatal(err.Error())
	}
}

func setupRouter(cfg Config) (*gin.Engine, error) {
	r := gin.Default()
	// TODO: Set mode to 'release' before deployment
	err := r.SetTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}
	r.POST("/message", func(c *gin.Context) {
		// TODO: EOF error is returned from the API when no body is given
		var req struct {
			Timestamp time.Time `json:"timestamp" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
			Value     string    `json:"value"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, req)
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "{'status':'OK'}")
	})
	return r, nil
}
