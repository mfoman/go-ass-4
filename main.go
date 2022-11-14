package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	api "github.com/mfoman/go-ass-4/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type STATE int64

const (
	RELEASED STATE = iota
	HELD
	WANTED
)

var port = flag.Int("port", 5000, "port")

type peer struct {
	api.UnimplementedDistributedMutualExclusionServer
	id           uint32
	clients      map[uint32]api.DistributedMutualExclusionClient
	defered      []uint32
	ctx          context.Context
	state        STATE
	clock        uint32
	requestsSent int
}

func (p *peer) Request(ctx context.Context, req *api.Request) (*api.RequestBack, error) {
	/*
		I received a request from a peer

		if my state is HELD or (WANTED and I have higher priority than my peer)
			then i defer with a REPLY until later(queue)
		else
			I send them a REPLY right away
	*/

	log.Printf("Got request from %v with clock %v\n", req.Id, req.Clock)

	if p.state == HELD || (p.state == WANTED && (p.clock < req.Clock || (p.clock == req.Clock && p.id > req.Id))) {
		log.Printf("Deferring request from %v with clock %v\n", req.Id, req.Clock)
		p.defered = append(p.defered, req.Id)
	} else {
		if p.state == WANTED {
			p.requestsSent++
			p.clients[req.Id].Request(ctx, &api.Request{Id: p.id, Clock: p.clock})
		}
		log.Printf("Sending reply right away to %v with clock %v\n", req.Id, req.Clock)
		p.clients[req.Id].Reply(ctx, &api.Reply{})
	}

	rep := &api.RequestBack{}
	return rep, nil
}

func (p *peer) Reply(ctx context.Context, req *api.Reply) (*api.ReplyBack, error) {
	/*
		I received REPLY from other peer

		if i received N-1 REPLIES(1 for each Request send) then
			i take CS(Critical Sec)
		else
			i wait for next reply
	*/

	p.requestsSent--

	if p.requestsSent == 0 {
		log.Println("Got a reply. All requests replied.")
		go p.criticalSection()
	}

	log.Printf("Got a reply. Missing %d replies.\n", p.requestsSent)

	rep := &api.ReplyBack{}
	return rep, nil
}

func (p *peer) criticalSection() {
	/*
		I have critical section
		When done I call sendReplyToAllDefered
	*/

	log.Println("Critical Section: Started working")
	p.state = HELD
	time.Sleep(5 * time.Second)
	p.state = RELEASED
	log.Println("Critical Section: Done working")

	p.sendReplyToAllDefered()
}

func (p *peer) sendRequestToAll() {
	p.state = WANTED

	request := &api.Request{
		Id:    p.id,
		Clock: p.clock,
	}

	p.requestsSent = len(p.clients)

	log.Printf("Sending request to all peers. Missing %d replies.\n", p.requestsSent)

	for id, client := range p.clients {
		_, err := client.Request(p.ctx, request)

		if err != nil {
			log.Printf("Something went wrong with %v\n", id)
		}
	}
}

func (p *peer) sendReplyToAllDefered() {
	reply := &api.Reply{}

	/*
		Reply all defered

		then clear p.defered
	*/

	for _, id := range p.defered {
		_, err := p.clients[id].Reply(p.ctx, reply)

		if err != nil {
			log.Printf("Something went wrong with %v\n", id)
		}
	}

	p.defered = make([]uint32, 0)
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var port = uint32(*port)

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

	p := &peer{
		id:           port,
		clients:      make(map[uint32]api.DistributedMutualExclusionClient),
		defered:      make([]uint32, 0),
		ctx:          ctx,
		state:        RELEASED,
		clock:        0,
		requestsSent: 0,
	}

	log.Printf("Opened on port: %d\n", p.id)

	grpcServer := grpc.NewServer()
	api.RegisterDistributedMutualExclusionServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		peerPort := uint32(5000 + i)

		if peerPort == p.id {
			continue
		}

		var conn *grpc.ClientConn
		log.Printf("Trying to dial: %v\n", peerPort)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", peerPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}

		defer conn.Close()
		log.Printf("Succes connecting to: %v\n", peerPort)
		c := api.NewDistributedMutualExclusionClient(conn)
		p.clients[peerPort] = c
	}

	log.Printf("I am connected to %d clients", len(p.clients))

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		log.Println("Requesting to work in critical section...")
		p.sendRequestToAll()
	}
}
