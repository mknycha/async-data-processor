package messagesconsumer

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

type acknowledgerMock struct {
}

func (m *acknowledgerMock) Ack(tag uint64, multiple bool) error                { return nil }
func (m *acknowledgerMock) Nack(tag uint64, multiple bool, requeue bool) error { return nil }
func (m *acknowledgerMock) Reject(tag uint64, requeue bool) error              { return nil }

func Test_worker_run(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		buff := &messagesBuffer{}
		outputFilenameSuffixFn := func() int64 { return 1 }
		w := newWorker("worker-1", buff, "./testdata", outputFilenameSuffixFn)
		msgsChan := make(chan amqp091.Delivery)
		ackMock := &acknowledgerMock{}
		var wg sync.WaitGroup
		wg.Add(1)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		go w.run(ctx, msgsChan, &wg)
		go func() {
			msgsChan <- amqp091.Delivery{
				Acknowledger: ackMock,
				MessageId:    "message-1",
				Body:         []byte("this is a test message"),
			}
			msgsChan <- amqp091.Delivery{
				Acknowledger: ackMock,
				MessageId:    "message-2",
				Body:         []byte("this is also a test message"),
			}
		}()
		wg.Wait()
		assert.FileExists(t, "./testdata/worker-1_1.txt")
		content, err := os.ReadFile("./testdata/worker-1_1.txt")
		if err != nil {
			t.Fatalf("failed to read output testfile: %s", err.Error())
		}
		assert.Contains(t, string(content), "this is a test message")
		assert.Contains(t, string(content), "this is also a test message")

		t.Cleanup(func() {
			err := os.Remove("./testdata/worker-1_1.txt")
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}
