// ========================= chat-service/main.go =========================
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"chat-service/shared"

	"github.com/redis/go-redis/v9"
)

var rdb = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

func main() {
	sub := rdb.Subscribe(context.Background(), "chat")
	ch := sub.Channel()

	fmt.Println("Chat Service listening to 'chat' channel")
	for msg := range ch {
		var payload shared.ChatPayload
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Println("Invalid payload")
			continue
		}

		fmt.Printf("[DB] Room %s | %s: %s\n", payload.Room, payload.Sender, payload.Message)

		event := shared.Message{
			Type:    "chat.message",
			Payload: json.RawMessage(msg.Payload),
		}
		out, _ := json.Marshal(event)
		rdb.Publish(context.Background(), "broadcast:"+payload.Room, string(out))
	}
}
