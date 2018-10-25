package main

import (
	"github.com/weAutomateEverything/requestReplyHal/requestReply"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {

	AlertRequestReplyService := requestReply.NewGrpcService()
	grpc := grpc.NewServer()
	reflection.Register(grpc)

	requestReply.RegisterAlertServer(grpc, AlertRequestReplyService)
	errs := make(chan error, 2)
	ln, err := net.Listen("tcp", ":50500")
	if err != nil {
		log.Println("transport", "grpc", "address", ":50500", "error", err)
		errs <- err
		panic(err)
	}
	go func() {
		log.Println("transport", "http", "address", ":50500", "msg", "listening")
		errs <- grpc.Serve(ln)
	}()
	log.Println("terminated", <-errs)

}
