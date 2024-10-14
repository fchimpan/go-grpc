package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	hellopb "github.com/fchimpan/go-gprc/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	scan   *bufio.Scanner
	client hellopb.GreetingServiceClient
)

func main() {
	log.Println("Client is running...")

	scan = bufio.NewScanner(os.Stdin)

	addr := "localhost:8081"
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	client = hellopb.NewGreetingServiceClient(conn)

	for {
		fmt.Println("1: send Request")
		fmt.Println("2: send Request to Stream")
		fmt.Println("3: send Request to Bi-Stream")
		fmt.Println("9: exit")
		fmt.Print("please enter >")

		scan.Scan()
		in := scan.Text()

		switch in {
		case "1":
			fmt.Println("Please enter your name.")
			scan.Scan()
			name := scan.Text()

			req := &hellopb.HelloRequest{
				Name: name,
			}
			res, err := client.SayHello(context.Background(), req)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(res.GetMessage())
			}

		case "2":
			fmt.Println("Please enter your name.")
			scan.Scan()
			name := scan.Text()

			req := &hellopb.HelloRequest{
				Name: name,
			}
			stream, err := client.SayHelloStream(context.Background(), req)
			if err != nil {
				fmt.Println(err)
				return
			}

			for {
				res, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					fmt.Println("all the responses have already received.")
					break
				}

				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(res)
			}
		case "3":
			stream, err := client.SayHelloBiStreams(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}

			sendNum := 5
			fmt.Printf("Please enter %d names.\n", sendNum)

			var sendEnd, recvEnd bool
			sendCount := 0
			for !(sendEnd && recvEnd) {
				// 送信処理
				if !sendEnd {
					scan.Scan()
					name := scan.Text()

					sendCount++
					if err := stream.Send(&hellopb.HelloRequest{
						Name: name,
					}); err != nil {
						fmt.Println(err)
						sendEnd = true
					}

					if sendCount == sendNum {
						sendEnd = true
						if err := stream.CloseSend(); err != nil {
							fmt.Println(err)
						}
					}
				}

				// 受信処理
				if !recvEnd {
					if res, err := stream.Recv(); err != nil {
						if !errors.Is(err, io.EOF) {
							fmt.Println(err)
						}
						recvEnd = true
					} else {
						fmt.Println(res.GetMessage())
					}
				}
			}
		case "9":
			fmt.Println("bye.")
			goto M
		}
	}
M:
}
