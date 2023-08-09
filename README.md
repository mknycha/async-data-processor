# Async web server

Demo for web server using pub/sub and background processing written in Go.

The app consists of two commands:
```
  message-consumer Runs the pool of workers that consumes messages from pub/sub and processes them
  server           Runs the web server that exposes an api for creating messages
```

## Server
The server exposes a single API endpoint `POST /message` and expects JSON.
It simply shards the data and sends it to pub/sub.
Example request body:
```
[
    {"timestamp":"2019-10-12T08:10:53.52Z","value":"D"},
    {"timestamp":"2019-10-12T08:10:52.52Z","value":"E"},
    {"timestamp":"2019-10-12T08:10:54.52Z","value":"F"}
]
```
Only timestamp is required for each entry.
Status 201 means the messages were sent to the queue.

## Message consumer
- Each shard has a separate worker pool (fixed number of workers for now)
- Worker processing timeout
- Any worker will buffer the data they receive and write the messages in batches in files (containing the worker id)
- Displays a history of workers and how many messages each has processed in the logs

## Prerequisites
- RabbitMQ 3.12.2
- Go 1.20
- Docker can be helpful

## How to run it:
1. Run rabbitMQ. Recommended docker: `docker run --rm --hostname my-rabbit -p 5672:5672 --name some-rabbit rabbitmq:3`
2. Create a directory that will be output for the processed data and set it to the env variable `DATA_OUTPUT_DIR`
3. Run `go build`
4. Run `async-data-processor server` and `async-data-processor message-consumer` in a separate console window

Assumptions:
- Sharding:
  - It is ok to shard by timestamp, because timestamp data is distributed more or less uniformly
- Regarding request:
  - Value can consist of several characters and it's not required
- Regarding sever:
  - It's ok to show some internal messages in the API response when internal server error is returned
- Regarding workers:
  - We want one of the workers in the pool to process five messages as soon as we have five messages in the queue. That's why workers in a pool are sharing the same batch. I know that batch can become a deep throat here, but I felt it's more reasonable than each worker having it's own batch and compete for unbatched messages with other workers.
  - Regarding the requirement: _Displays a history of workers and how many messages each has processed_ I assumed that console logs are fine.

TODO/Corners cut:
- There is no graceful shutdown on termination signal. That is, CTRL+C can cause some of the messages to be unprocessed (those that are in the batch).
- Requirement _It starts with 3 workers and has at most 4 workers at any time_ **is not fullfilled**, my worker pool spawns fixed number of 3 workers (configurable).
- Very naive error handling in the worker (I would use errgroup if I had more time)
- Messages buffer could be refactored, `GetMessages` and `Reset` should be in one function
- Test coverage is low