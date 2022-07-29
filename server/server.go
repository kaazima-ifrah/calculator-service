package main

import (
	"context"
	"fmt"
	"github.com/bootcamp/calculator-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"time"
)

type server struct {
	api.UnimplementedCalculatorServiceServer
}

func (*server) Sum(ctx context.Context, request *api.SumRequest) (response *api.SumResponse, err error) {
	fmt.Println("Starting unary streaming to find the sum of 2 numbers ...")

	num1 := request.GetNum1()
	num2 := request.GetNum2()

	result := num1 + num2
	response = &api.SumResponse{
		Sum: result,
	}
	return response, nil
}

func (*server) PrimeNumbers(request *api.PrimeNumbersRequest, response api.CalculatorService_PrimeNumbersServer) error {
	fmt.Println("Starting server side streaming to list the prime numbers ...")

	num := request.GetNum()

	isPrime := func(num int) bool {
		for i := 2; i < num; i++ {
			if num%i == 0 {
				return false
			}
		}
		return true
	}

	for i := 2; i < int(num); i++ {
		if isPrime(i) {
			result := api.PrimeNumbersResponse{
				PrimeNumber: int32(i),
			}
			time.Sleep(500 * time.Millisecond)
			sendErr := response.Send(&result)
			if sendErr != nil {
				log.Fatalf("Error while sending response from PrimeNumbersServer: %v", sendErr)
				return sendErr
			}
		}
	}
	return nil
}

func (*server) ComputeAverage(request api.CalculatorService_ComputeAverageServer) error {
	fmt.Println("Starting client side streaming to calculate the average ...")

	var sum, count int32 = 0, 0
	for {
		req, err := request.Recv()
		if err == io.EOF {
			// finished reading client stream
			return request.SendAndClose(&api.ComputeAverageResponse{
				Average: float32(sum) / float32(count),
			})
		}
		if err != nil {
			log.Fatalf("Error while reading ComputeAverageClient stream : %v", err)
			return err
		}
		sum += req.GetNum()
		count += 1
	}
}

func (*server) MaxNumber(request api.CalculatorService_MaxNumberServer) error {
	fmt.Println("Starting bi directional streaming to find the max number ...")

	var max int32 = 0
	for {
		req, err := request.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while receiving data from MaxNumberClient : %v", err)
			return err
		}

		if req.GetNum() > max {
			max = req.Num
			sendErr := request.Send(&api.MaxNumberResponse{
				MaxNumber: max,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending response from MaxNumberServer: %v", sendErr)
				return sendErr
			}
		}
	}
}

func main() {
	fmt.Println("Starting server on port 50001")

	listen, err := net.Listen("tcp", "0.0.0.0:50001")
	if err != nil {
		log.Fatalf("Failed to Listen: %v", err)
	}

	// create a gRPC server object
	grpcServer := grpc.NewServer()

	// attach service to server
	api.RegisterCalculatorServiceServer(grpcServer, &server{})

	// register reflection service on gRPC server
	reflection.Register(grpcServer)

	// start the server
	if err = grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve : %v", err)
	}
}
