package requestReply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/weAutomateEverything/go2hal/alert"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func NewGrpcService() AlertServer {
	s := &grpcService{
		requests: make(map[int64]Alert_RequestReplyServer),
	}
	go func() {
		s.poll()

	}()
	return s
}

type grpcService struct {
	requests map[int64]Alert_RequestReplyServer
}

func (s *grpcService) RequestReply(a Alert_RequestReplyServer) error {
	for {
		in, err := a.Recv()
		if err == io.EOF{
			return nil
		}
		if err != nil {
			sendError(fmt.Errorf("error receiving grpc alert: %v", err))
			return err
		}
		if in.Message != "" {
			_, err = sendMessage(in.GroupID, in.Message, in.CorrelationId)
			if err != nil {
				sendError(fmt.Errorf("error sending grpc alert message: %v", err))
				continue
			}
		}
		s.requests[in.GroupID] = a
	}
}

func (s grpcService) poll() {
	for {
		time.Sleep(2 * time.Second)
		requests:
		for id, server := range s.requests {
			replies, err := getReplies(id)
			if err != nil {
				sendError(fmt.Errorf("error polling for replies %v", err))
			}
			for _, reply := range replies {
				err = server.Send(
					&AlertReply{
						CorrelationId: reply.CorrelationId,
						Message:       reply.Message,
					},
				)

				if err != nil {
					delete(s.requests,id)
					sendError(fmt.Errorf("error sending grpc reply %v", err))
					continue requests
				}
				err = acknoweldgeReply(id, reply.MessageId)
				if err != nil {
					sendError(fmt.Errorf("error acknowledging reply %v", err))
					continue
				}
			}

		}

	}
}

func sendError(err error) {
	resp, err := http.Post(fmt.Sprintf("%v/error", getAlertEndpoint()), "application/text", strings.NewReader(err.Error()))
	if err != nil {
		log.Println(err)
		return
	}
	resp.Body.Close()
}

func sendMessage(group int64, msg string, correlation string) (id string, err error) {
	r := alert.SendReplyAlertMessageRequest{
		CorrelationId: correlation,
		Message:       msg,
	}
	m, err := json.Marshal(r)
	if err != nil {
		return
	}
	resp, err := http.Post(fmt.Sprintf("%v/%v/withreply", getAlertEndpoint(), group), "application/json", bytes.NewReader(m))
	if err != nil {
		return
	}
	resp.Body.Close()
	return

}

func getReplies(id int64) (r []alert.Reply, err error) {
	resp, err := http.Get(fmt.Sprintf("%v/%v/replies", getAlertEndpoint(), id))
	if err != nil {
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&r)
	resp.Body.Close()
	return
}

func acknoweldgeReply(chat int64, id string) (err error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/%v/reply/%v", getAlertEndpoint(), chat, id), nil)
	if err != nil {
		return
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	r.Body.Close()
	return
}

func getAlertEndpoint() string {
	return fmt.Sprintf("%v/api/alert", os.Getenv("HAL_ENDPOINT"))
}
