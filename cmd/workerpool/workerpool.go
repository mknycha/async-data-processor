package workerpool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mknycha/async-data-processor/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

type Config struct {
	RabbitmqUrl            string `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`
	ShardsCount            int    `envconfig:"SHARDS_COUNT" default:"5"`
	WorkersCount           int    `envconfig:"WORKERS_COUNT" default:"1"`
	WorkersWorktimeSeconds int    `envconfig:"WORKERS_WORKTIME_SECONDS" default:"120"`
}

func WorkerPoolCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "worker-pool",
		Short: "Runs the pool of workers that consumes and processes messages from pub/sub",
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

			var wg sync.WaitGroup
			for i := 0; i < cfg.ShardsCount; i++ {
				err := wrapper.QueueDeclare(i)
				if err != nil {
					log.Fatalf("failed to declare a queue: %s", err.Error())
				}
				msgs, err := wrapper.MessagesChannel(i)
				if err != nil {
					log.Fatalf("failed to get messages channel: %s", err.Error())
				}
				for j := 0; j < cfg.WorkersCount; j++ {
					wg.Add(1)
					workerId := fmt.Sprintf("%d-%d", i, j)
					log.Printf("spawning worker #%s\n", workerId)
					ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(cfg.WorkersWorktimeSeconds)*time.Second)
					defer cancelFunc()
					go spawnWorker(ctx, msgs, &wg, workerId)
				}
			}

			// log.Printf("Waiting for messages. To exit press CTRL+C")

			// c := make(chan os.Signal, 1)
			// signal.Notify(c, os.Interrupt)
			// <-c
			// cancelFunc()
			wg.Wait()
			fmt.Println("goodbye!")
		},
	}
}

func spawnWorker(ctx context.Context, msgs <-chan amqp.Delivery, wg *sync.WaitGroup, workerId string) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("stopping worker #%s...\n", workerId)
			wg.Done()
			return
		// TODO: What if the msgs channel is closed?
		case msg := <-msgs:
			log.Printf("Worker #%s received a message: %s", workerId, string(msg.Body))
			msg.Ack(false)
		}
	}
}
