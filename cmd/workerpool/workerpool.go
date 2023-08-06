package workerpool

import (
	"log"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/mknycha/async-data-processor/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

type Config struct {
	RabbitmqUrl string `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`
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
			log.Printf("Waiting for messages. To exit press CTRL+C")
			msgs, err := wrapper.MessagesChannel()
			if err != nil {
				log.Fatalf("failed to consume messages: %s", err.Error())
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			for {
				select {
				case <-c:
					log.Println("Goodbye!")
					return
				case d := <-msgs:
					log.Printf("Received a message: %s", d.Body)
					d.Ack(false)
				}
			}
		},
	}
}
