package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	ping "github.com/mfoman/go-ass-4/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var port = flag.Int("port", 5000, "port")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var port = uint32(*port)

	p := &peer{
		id:            port,
		amountOfPings: make(map[uint32]uint32),
		clients:       make(map[uint32]ping.PingClient),
		ctx:           ctx,
	}

	// Create listener tcp on port ownPort
	var attempts = 0

	var list net.Listener

	for {
		var err error
		list, err = net.Listen("tcp", fmt.Sprintf(":%v", port))

		if err != nil {
			if attempts > 5 {
				log.Fatalf("Failed to listen on port: %v - with %d attempts", err, attempts)
			}

			attempts++
			port++

			continue
		}

		break
	}

	log.Printf("Opened on port: %d\n", port)

	grpcServer := grpc.NewServer()
	ping.RegisterPingServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		peerPort := uint32(5000 + i)

		if peerPort == port {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", peerPort)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", peerPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}

		defer conn.Close()
		c := ping.NewPingClient(conn)
		p.clients[peerPort] = c
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		log.Println("ping all")
		p.sendPingToAll()
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id            uint32
	amountOfPings map[uint32]uint32
	clients       map[uint32]ping.PingClient
	ctx           context.Context
}

func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	id := req.Id
	p.amountOfPings[id] += 1

	rep := &ping.Reply{Amount: p.amountOfPings[id]}
	return rep, nil
}

func (p *peer) sendPingToAll() {
	request := &ping.Request{Id: p.id}

	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)

		if err != nil {
			fmt.Println("something went wrong")
		}

		fmt.Printf("Got reply from id %v: %v\n", id, reply.Amount)
	}
}
