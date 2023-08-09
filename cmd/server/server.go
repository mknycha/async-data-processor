package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/mknycha/async-data-processor/pubsub"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

const (
	timeout = time.Second * 5
)

//go:generate mockgen -destination mocks/generated.go --package mocks --source server.go
type wrapper interface {
	PublishWithContext(ctx context.Context, messageBody []byte, shardNo int) error
}

type Config struct {
	TrustedProxies []string `envconfig:"TRUSTED_PROXIES"`
	Port           string   `envconfig:"PORT" default:"8080"`
	RabbitmqUrl    string   `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`
	ShardsCount    int      `envconfig:"SHARDS_COUNT" default:"5"`
}

func ServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Runs the web server that exposes an api for creating messages",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg Config
			err := envconfig.Process("", &cfg)
			if err != nil {
				log.Fatalf("failed to process config: %s", err.Error())
			}

			conn, err := amqp.Dial(cfg.RabbitmqUrl)
			if err != nil {
				log.Fatalf("failed to initialize rabbitmq connection: %s", err.Error())
			}
			defer conn.Close()

			wrapper, err := pubsub.NewWrapper(conn)
			if err != nil {
				log.Fatalf("failed to initialize wrapper: %s", err.Error())
			}
			for i := 0; i < cfg.ShardsCount; i++ {
				err = wrapper.QueueDeclare(i)
				if err != nil {
					log.Fatalf("failed to declare a queue: %s", err.Error())
				}
			}

			router, err := setupRouter(cfg, wrapper)
			if err != nil {
				log.Fatal(err)
			}
			err = router.Run(fmt.Sprintf(":%s", cfg.Port))
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
}

type Entry struct {
	Timestamp time.Time `json:"timestamp" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	Value     string    `json:"value"`
}

func setupRouter(cfg Config, wrapper wrapper) (*gin.Engine, error) {
	r := gin.Default()
	// TODO: Set mode to 'release' before deployment
	err := r.SetTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}
	r.POST("/message", func(c *gin.Context) {
		// TODO: EOF error is returned from the API when no body is given
		var req []Entry
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, ent := range req {
			// TODO: Can it happen in parallel?
			err = sendEntryForProcessing(wrapper, ent, cfg.ShardsCount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusCreated, req)
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "{'status':'OK'}")
	})
	return r, nil
}

func sendEntryForProcessing(wrapper wrapper, ent Entry, shardsCount int) error {
	jsonContent, err := json.Marshal(ent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal json")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	shardNo := shardRouting(ent, shardsCount)
	err = wrapper.PublishWithContext(ctx, jsonContent, shardNo)
	if err != nil {
		return errors.Wrap(err, "failed to publish message")
	}
	log.Printf("Published message: %s to shard #%d\n", string(jsonContent), shardNo)
	return nil
}

func shardRouting(ent Entry, shardCount int) int {
	return int(ent.Timestamp.Unix()) % shardCount
}
