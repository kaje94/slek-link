
package asyncapi

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// GetRouter returns a watermill router. 
func GetRouter() (*message.Router, error){
	logger := watermill.NewStdLogger(false, false)
	return message.NewRouter(message.RouterConfig{}, logger)
}


// ConfigureAMQPSubscriptionHandlers configures the router with the subscription handler.    
func ConfigureAMQPSubscriptionHandlers(r *message.Router, s message.Subscriber) {

  r.AddNoPublisherHandler(
    "OnUrlVisited",     // handler name, must be unique
    "url/visited",         // topic from which we will read events
    s,
    OnUrlVisited, 
  )

  r.AddNoPublisherHandler(
    "PublishUrlVisitedError",     // handler name, must be unique
    "url/visited/error",         // topic from which we will read events
    s,
    PublishUrlVisitedError, 
  )

}    

  