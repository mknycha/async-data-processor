package messagesconsumer

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
	WorkersCount           int    `envconfig:"WORKERS_COUNT" default:"3"`
	WorkersWorktimeSeconds int    `envconfig:"WORKERS_WORKTIME_SECONDS" default:"120"`
	DataOutputDir          string `envconfig:"DATA_OUTPUT_DIR" default:"./data_dump"`
}

func MessageConsumerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "message-consumer",
		Short: "Runs the pool of workers that consumes messages from pub/sub and processes them",
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

			ctx := context.Background()
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
				batch := &messagesBuffer{}
				pool := newWorkerPool(i, batch, cfg.WorkersCount, cfg.DataOutputDir)
				ctx, cancelFunc := context.WithTimeout(ctx, time.Duration(cfg.WorkersWorktimeSeconds)*time.Second)
				defer cancelFunc()
				wg.Add(cfg.WorkersCount)
				pool.SpawnWorkers(ctx, msgs, &wg)
			}

			wg.Wait()
			fmt.Println("goodbye!")
		},
	}
}

type workerPool struct {
	id            int
	buff          *messagesBuffer
	workersCount  int
	dataOutputDir string
}

func newWorkerPool(id int, buff *messagesBuffer, workersCount int, dataOutputDir string) *workerPool {
	return &workerPool{
		id:            id,
		buff:          buff,
		workersCount:  workersCount,
		dataOutputDir: dataOutputDir,
	}
}

func (wp *workerPool) SpawnWorkers(ctx context.Context, msgs <-chan amqp.Delivery, wg *sync.WaitGroup) {
	for i := 0; i < wp.workersCount; i++ {
		workerId := fmt.Sprintf("%d-%d", wp.id, i)
		w := newWorker(workerId, wp.buff, wp.dataOutputDir, time.Now().UnixNano)
		go w.run(ctx, msgs, wg)
	}
}
