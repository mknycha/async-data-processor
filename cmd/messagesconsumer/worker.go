package messagesconsumer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const bufferSizeToFlush = 5

type worker struct {
	workerId               string
	buffer                 *messagesBuffer
	outputDir              string
	outputFilenameSuffixFn func() int64
	bufferSizeToFlush      int
}

func newWorker(workerId string, buffer *messagesBuffer, outputDir string, outputFilenameSuffixFn func() int64) *worker {
	return &worker{
		workerId:               workerId,
		buffer:                 buffer,
		outputDir:              outputDir,
		outputFilenameSuffixFn: outputFilenameSuffixFn,
		bufferSizeToFlush:      bufferSizeToFlush,
	}
}

func (w *worker) run(ctx context.Context, msgs <-chan amqp.Delivery, wg *sync.WaitGroup) {
	log.Printf("starting worker #%s\n", w.workerId)
	for {
		select {
		case <-ctx.Done():
			// flush the remaining messages before stopping
			w.flush(w.buffer)
			log.Printf("stopping worker #%s...\n", w.workerId)
			wg.Done()
			return
		case msg := <-msgs:
			w.buffer.Push(msg.Body)
			err := msg.Ack(false)
			if err != nil {
				log.Printf("error: failed to ack message: %s\n", msg.MessageId)
				return
			}
			if w.buffer.Size() == bufferSizeToFlush {
				w.flush(w.buffer)
			}
		}
	}
}

func (w *worker) flush(b *messagesBuffer) {
	batchedMessages := b.GetMessages()
	if len(batchedMessages) == 0 {
		return
	}
	b.Reset()
	content := bytes.Join(batchedMessages, []byte("\n"))
	filename := w.outputFilename()
	err := os.WriteFile(filename, content, 0644)
	if err != nil {
		log.Printf("error: failed to write file %q: %s\n", filename, err.Error())
		return
	}
	log.Printf("Worker #%s processed %d messages, wrote to file: %q", w.workerId, len(batchedMessages), filename)
}

func (w *worker) outputFilename() string {
	return fmt.Sprintf("%s/%s_%d.txt", w.outputDir, w.workerId, w.outputFilenameSuffixFn())
}
