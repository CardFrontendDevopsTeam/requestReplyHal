package main

import (
	"context"
	"github.com/weAutomateEverything/requestReplyHal/requestReply"
	"google.golang.org/grpc"
	"log"
)

func main() {
	a := requestReply.AlertRequest{
		Message:       "Hello World",
		CorrelationId: "12345",
		GroupID:       72893782,
	}

	con, err := grpc.Dial("grpc.hal.cloudy.standardbank.co.za:50500", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c := requestReply.NewAlertClient(con)

	client, err := c.RequestReply(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	err = client.Send(&a)
	if err != nil {
		log.Fatal(err)
	}
	for {
		reply, err := client.Recv()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v - %v", reply.CorrelationId, reply.Message)
	}
}
