package server

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gbh007/p2p-chat/internal/entities"
	"github.com/gbh007/p2p-chat/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	gen.UnimplementedServerServer
	logger *slog.Logger

	readers      map[string]map[string]chan entities.Message
	readersMutex *sync.RWMutex
}

func New() *Server {
	return &Server{
		readers:      make(map[string]map[string]chan entities.Message),
		readersMutex: &sync.RWMutex{},
		logger:       slog.Default(),
	}
}

func (s *Server) ReadMessages(req *gen.ReadMessagesRequest, stream grpc.ServerStreamingServer[gen.ReadMessagesResponse]) error {
	ctx := stream.Context()

	ch := make(chan entities.Message, 100)

	s.readersMutex.Lock()

	users, ok := s.readers[req.GetChannel()]
	if !ok {
		users = make(map[string]chan entities.Message)
		s.logger.Info("create chan", "chan", req.GetChannel())
		s.readers[req.GetChannel()] = users
	}

	_, ok = users[req.GetLogin()]
	if ok {
		s.readersMutex.Unlock()

		return fmt.Errorf("already connected")
	}

	users[req.GetLogin()] = ch
	s.logger.Info("create user listen", "chan", req.GetChannel(), "user", req.GetLogin())

	s.readersMutex.Unlock()

	var sendError error

listen:
	for {
		select {
		case msg := <-ch:
			sendError = stream.Send(&gen.ReadMessagesResponse{
				Login:   msg.User,
				Message: msg.Text,
				Ts:      timestamppb.New(msg.TS),
			})
			if sendError != nil {
				break listen
			}
		case <-ctx.Done():
			break listen
		}
	}

	s.readersMutex.Lock()
	delete(s.readers[req.GetChannel()], req.GetLogin())
	s.readersMutex.Unlock()

	close(ch)

	// Дочитываем чтобы разблокировать другие потоки
	for range ch {
	}

	if sendError != nil {
		return fmt.Errorf("send: %w", sendError)
	}

	return nil
}

func (s *Server) SendMessage(ctx context.Context, req *gen.SendMessageRequest) (*gen.SendMessageResponse, error) {
	s.readersMutex.RLock()
	defer s.readersMutex.RUnlock()

	users, ok := s.readers[req.GetChannel()]
	if !ok {
		s.logger.Info("missing channel", "chan", req.GetChannel(), "user", req.GetLogin())
		return nil, fmt.Errorf("missing channel")
	}

	msg := entities.Message{
		Chat: req.GetChannel(),
		User: req.GetLogin(),
		Text: req.GetMessage(),
		TS:   time.Now(),
	}

	for _, ch := range users {
		ch <- msg
	}

	return &gen.SendMessageResponse{
		Ts: timestamppb.Now(),
	}, nil
}
