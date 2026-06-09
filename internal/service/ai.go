package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"eino-stock/internal/biz/ai"
	"eino-stock/internal/infrastructure/eino"

	"github.com/cloudwego/eino/schema"
)

type AIService struct {
	uc *ai.ScreenUsecase
}

func NewAIService(uc *ai.ScreenUsecase) *AIService {
	return &AIService{uc: uc}
}

func (s *AIService) Screen(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" { writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter q")); return }
	result, err := s.uc.Screen(r.Context(), q)
	writeJSON(w, result, err)
}

func (s *AIService) ParallelScreen(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" { writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter q")); return }
	result, err := s.uc.ParallelScreen(r.Context(), q)
	writeJSON(w, result, err)
}

type streamMsg struct {
	msg *schema.Message
	err error
}

func (s *AIService) ChatStream(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" { writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter q")); return }

	// Custom prompt support
	if pt := r.URL.Query().Get("prompt_text"); pt != "" {
		q = pt + "\n\n请严格按照以下策略分析：\n" + q
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)

	send := func(t, c string) {
		fmt.Fprintf(w, "data: {\"type\":%q,\"content\":%q}\n\n", t, c)
		if flusher != nil { flusher.Flush() }
	}

	send("start", "")
	send("status", "正在初始化AI...")

	cfg := eino.ReadAIConfig()
	agent, err := eino.NewChatAgent(r.Context(), cfg)
	if err != nil { send("error", err.Error()); send("done", ""); return }

	send("status", "AI初始化完成，正在分析...")

	stream, err := agent.Stream(r.Context(), q, func(name, args string) {
		fmt.Fprintf(w, "data: {\"type\":\"tool\",\"name\":%q,\"content\":\"正在查询...\"}\n\n", name)
		if flusher != nil { flusher.Flush() }
	})
	if err != nil { send("error", err.Error()); send("done", ""); return }
	defer stream.Close()

	ch := make(chan streamMsg, 1)
	go func() { m, e := stream.Recv(); ch <- streamMsg{m, e} }()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				if r.err != io.EOF {
					send("error", r.err.Error())
				}
				send("done", "")
				return
			}
			if r.msg != nil && r.msg.Content != "" {
				c := strings.ReplaceAll(r.msg.Content, "\n", "\\n")
				c = strings.ReplaceAll(c, "\"", "\\\"")
				send("text", c)
			}
			go func() { m, e := stream.Recv(); ch <- streamMsg{m, e} }()
		case <-ticker.C:
			send("ping", "thinking")
		case <-r.Context().Done():
			send("done", "")
			return
		}
	}
}
