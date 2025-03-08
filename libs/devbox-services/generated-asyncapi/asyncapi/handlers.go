
package asyncapi

import (
	"log"
  
  "encoding/json"
  "github.com/ThreeDotsLabs/watermill/message"

  "context"
  "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
)

// OnUrlVisited subscription handler for url/visited.
func OnUrlVisited(msg *message.Message) error {
    log.Printf("received message payload: %s", string(msg.Payload))

    var lm UrlVisitedPayload
    err := json.Unmarshal(msg.Payload, &lm)
    if err != nil {
        log.Printf("error unmarshalling message: %s, err is: %s", msg.Payload, err)
    }
    return nil
}

// PublishUrlVisitedError subscription handler for error_queue.
func PublishUrlVisitedError(msg *message.Message) error {
    log.Printf("received message payload: %s", string(msg.Payload))

    var lm UrlVisitedPayload
    err := json.Unmarshal(msg.Payload, &lm)
    if err != nil {
        log.Printf("error unmarshalling message: %s, err is: %s", msg.Payload, err)
    }
    return nil
}


// PublishUrlVisited is the publish handler for url/visited.
func PublishUrlVisited(ctx context.Context, a *amqp.Publisher, payload UrlVisitedPayload) error {
  m, err := PayloadToMessage(payload)
  if err != nil {
      log.Fatalf("error converting payload: %+v to message error: %s", payload, err)
  }

  return a.Publish("url/visited", m)
}

