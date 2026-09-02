package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/Bambi-uxx/Meowmico/backend/db"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func getHistory() []GroqMessage {
	rows, err := db.DB.Query(
		"SELECT role, content FROM messages ORDER BY created_at DESC LIMIT 10",
	)
	if err != nil {
		return []GroqMessage{}
	}
	defer rows.Close()

	var messages []GroqMessage
	for rows.Next() {
		var m GroqMessage
		if err := rows.Scan(&m.Role, &m.Content); err != nil {
			continue
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return []GroqMessage{}
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages
}

func Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db.DB.Exec("INSERT INTO messages (role, content) VALUES (?, ?)", "user", req.Message)

	promptBytes, err := os.ReadFile("prompts/meowmico.md")
	if err != nil {
		promptBytes = []byte("You are Meowmico, a sarcastic pixel cat. Be funny and brief.")
	}
	systemPrompt := string(promptBytes)

	apiKey := os.Getenv("GROQ_API_KEY")

	var reply string

	if apiKey == "" {
		fallbacks := []string{
			"No API key? Really? I can't even. *sits on keyboard*",
			"You forgot the API key. Classic human behavior.",
			"I would respond but someone forgot to configure me. Not naming names.",
		}
		reply = fallbacks[len(req.Message)%len(fallbacks)]
	} else {
		history := getHistory()

		allMessages := []GroqMessage{
			{Role: "system", Content: systemPrompt},
		}
		for _, m := range history {
			allMessages = append(allMessages, m)
		}
		allMessages = append(allMessages, GroqMessage{
			Role:    "user",
			Content: req.Message,
		})

		groqReq := GroqRequest{
			Model:    "llama-3.1-8b-instant",
			Messages: allMessages,
		}

		body, _ := json.Marshal(groqReq)
		httpReq, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			reply = "My connection to the universe is broken. Try again."
		} else {
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			var groqResp GroqResponse
			if err := json.Unmarshal(respBody, &groqResp); err != nil || len(groqResp.Choices) == 0 {
				reply = "Something went wrong. Very on brand for this setup."
			} else {
				reply = groqResp.Choices[0].Message.Content
			}
		}
	}

	db.DB.Exec("INSERT INTO messages (role, content) VALUES (?, ?)", "meowmico", reply)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Reply: reply})
}
