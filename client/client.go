package main

import (
	"context"
	"fmt"
	"github.com/bootcamp/calculator-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"time"
)

func Sum(client api.CalculatorServiceClient) {
	fmt.Println("Sending unary request to find the sum of 2 numbers ...")

	request := api.SumRequest{
		Num1: 10,
		Num2: 23,
	}
	response, err := client.Sum(context.Background(), &request)
	if err != nil {
		log.Fatalf("Error while calling sum grpc unary call: %v", err)
	}

	fmt.Println("Response from Sum Unary Call : ", response.Sum)
}

func PrimeNumbers(client api.CalculatorServiceClient) {
	fmt.Println("Sending server side streaming request to list the prime numbers ...")

	request := api.PrimeNumbersRequest{
		Num: 26,
	}
	respStream, err := client.PrimeNumbers(context.Background(), &request)
	if err != nil {
		log.Fatalf("Error while calling PrimeNumbers server-side streaming grpc : %v", err)
	}
	for {
		msg, err := respStream.Recv()
		if err == io.EOF {
			//we have reached to the end of the file
			break
		}
		if err != nil {
			log.Fatalf("Error while receving server stream : %v", err)
		}
		fmt.Println("Response From PrimeNumbers Server : ", msg.PrimeNumber)
	}
}

func ComputeAverage(client api.CalculatorServiceClient) {
	fmt.Println("Sending client side streaming request to get the average ...")

	stream, err := client.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error occured while performing client-side streaming : %v", err)
	}

	requests := []*api.ComputeAverageRequest{
		{Num: 3},
		{Num: 12},
		{Num: 20},
	}

	for _, req := range requests {
		fmt.Println("\nSending Request : ", req)
		sendErr := stream.Send(req)
		if sendErr != nil {
			log.Fatalf("Error while sending request from ComputeAverageClient: %v", sendErr)
			return
		}
		time.Sleep(1 * time.Second)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from ComputeAverage server : %v", err)
	}
	fmt.Println("\nResponse From ComputeAverage Server : ", response.GetAverage())
}

func MaxNumber(client api.CalculatorServiceClient) {
	fmt.Println("Sending bi directional streaming request to get the max number ...")

	requests := []*api.MaxNumberRequest{
		{Num: 13},
		{Num: 12},
		{Num: 20},
		{Num: 49},
		{Num: 6},
	}

	stream, err := client.MaxNumber(context.Background())
	if err != nil {
		log.Fatalf("Error occured while performing bi directional streaming : %v", err)
	}

	// wait channel to block receiver
	waitchan := make(chan struct{})

	go func(requests []*api.MaxNumberRequest) {
		for _, req := range requests {
			fmt.Println("\nSending Request : ", req.GetNum())
			sendErr := stream.Send(req)
			if sendErr != nil {
				log.Fatalf("Error while sending request to MaxNumber service : %v", err)
			}
			time.Sleep(1000 * time.Millisecond)
		}
		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Error while closing the request to MaxNumber service : %v", err)
			return
		}
	}(requests)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitchan)
				return
			}
			if err != nil {
				log.Fatalf("Error receiving response from server : %v", err)
			}
			fmt.Println("\nResponse From MaxNumber Server : ", resp.GetMaxNumber())
		}
	}()

	//block until everything is finished
	<-waitchan
}

func main() {
	cc, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer func(cc *grpc.ClientConn) {
		err := cc.Close()
		if err != nil {
			log.Fatalf("could not close the client connection: %v", err)
		}
	}(cc)

	client := api.NewCalculatorServiceClient(cc)

	// Unary function
	// Sum(client)

	// Server-Side Streaming function
	// PrimeNumbers(client)

	// Client-Side Streaming function
	// ComputeAverage(client)

	// bidirectional streaming function
	MaxNumber(client)
}
