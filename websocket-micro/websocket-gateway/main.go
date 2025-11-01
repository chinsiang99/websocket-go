package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"websocket-gateway/shared"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var rdb = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

var secret = []byte("supersecret")

func authenticate(r *http.Request) (string, error) {
	tokenStr := r.URL.Query().Get("token")
	fmt.Println(tokenStr, "this is token string")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		log.Println("yo it goes over here man")
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	userID, ok := claims["userId"].(string)
	if !ok {
		return "", fmt.Errorf("userId not found")
	}
	return userID, nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	room := r.URL.Query().Get("room")
	if room == "" {
		room = "general"
	}

	sub := rdb.Subscribe(context.Background(), "broadcast:"+room)
	ch := sub.Channel()

	joinMsg := shared.ChatPayload{
		Room:    room,
		Sender:  userID,
		Message: fmt.Sprintf("%s has joined the room", userID),
	}
	joinData, _ := json.Marshal(joinMsg)
	rdb.Publish(context.Background(), "chat", joinData)

	go func() {
		for msg := range ch {
			conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			leaveMsg := shared.ChatPayload{
				Room:    room,
				Sender:  userID,
				Message: fmt.Sprintf("%s has left the room", userID),
			}
			leaveData, _ := json.Marshal(leaveMsg)
			rdb.Publish(context.Background(), "chat", leaveData)
			log.Println("Read error:", err)
			return
		}

		var incoming shared.Message
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("Invalid message format")
			continue
		}

		if incoming.Type == "chat.message" {
			var payload shared.ChatPayload
			if err := json.Unmarshal(incoming.Payload, &payload); err != nil {
				log.Println("Invalid payload")
				continue
			}
			payload.Sender = userID
			fixed, _ := json.Marshal(payload)
			rdb.Publish(context.Background(), "chat", string(fixed))
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("WebSocket Gateway running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
