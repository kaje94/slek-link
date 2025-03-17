
  package asyncapi

  import (
    "encoding/json"

    "github.com/ThreeDotsLabs/watermill/message"
  )
  
    
    // UrlVisitedPayload represents a UrlVisitedPayload model.
type UrlVisitedPayload struct {
  LinkId string `json:"linkId"`
  Timestamp string `json:"timestamp"`
  CountryCode string `json:"countryCode"`
  AdditionalProperties map[string]interface{} `json:"additionalProperties"`
}
    

// PayloadToMessage converts a payload to watermill message
func PayloadToMessage(i interface{}) (*message.Message, error) {
  var m message.Message

  b, err := json.Marshal(i)
  if err != nil {
    return nil, nil
  }
  m.Payload = b

  return &m, nil
}
  