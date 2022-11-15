package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
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
	mu           sync.Mutex
}

func (p *peer) incrementClock() {
	p.mu.Lock()
	p.clock++
	p.mu.Unlock()
}

func LamportMax(a, b uint32) uint32 {
	if a > b {
		return a + 1
	}
	return b + 1
}

func (p *peer) Request(ctx context.Context, req *api.Request) (*api.RequestBack, error) {
	/*
		I received a request from a peer

		if my state is HELD or (WANTED and I have higher priority than my peer)
			then i defer with a REPLY until later(queue)
		else
			I send them a REPLY right away
	*/

	log.Printf("(L: %d) RECV: Received request from %d\n", p.clock, req.Id)
	if p.state == HELD || (p.state == WANTED && (p.clock < req.Clock || (p.clock == req.Clock && p.id > req.Id))) {
		log.Printf("(L: %d) Deferring request from %v\n", p.clock, req.Id)
		p.defered = append(p.defered, req.Id)

		p.mu.Lock()
		p.clock = LamportMax(p.clock, req.Clock)
		p.mu.Unlock()
	} else {
		if p.state == WANTED {
			p.requestsSent++
			p.incrementClock()
			log.Printf("(L: %d) SEND: Rerequesting critical section to higher priority peer\n", p.clock)
			p.clients[req.Id].Request(ctx, &api.Request{Id: p.id, Clock: p.clock})
		}
		p.incrementClock()
		log.Printf("(L: %d) SEND: Sending reply right away to %v\n", p.clock, req.Id)
		p.clients[req.Id].Reply(ctx, &api.Reply{
			Clock: p.clock,
			Id:    p.id,
		})
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

	p.mu.Lock()
	p.clock = LamportMax(p.clock, req.Clock)
	p.mu.Unlock()

	p.requestsSent--
	log.Printf("(L: %d) RECV: Got a reply. Missing %d replies.\n", p.clock, p.requestsSent)

	if p.requestsSent == 0 {
		log.Printf("(L: %d) RECV: All requests replied.\n", p.clock)
		go p.criticalSection()
	}

	rep := &api.ReplyBack{}
	return rep, nil
}

func (p *peer) criticalSection() {
	/*
		I have critical section
		When done I call sendReplyToAllDefered
	*/

	p.incrementClock()
	log.Printf("(L: %d) CS: Started working\n", p.clock)
	p.state = HELD
	time.Sleep(5 * time.Second)
	p.state = RELEASED
	p.incrementClock()
	log.Printf("(L: %d) CS: Done working\n", p.clock)

	p.sendReplyToAllDefered()
}

func (p *peer) sendRequestToAll() {
	p.state = WANTED

	p.incrementClock()

	request := &api.Request{
		Id:    p.id,
		Clock: p.clock,
	}

	p.requestsSent = len(p.clients)

	log.Printf("(L: %d) SEND: Sending request to all peers. Missing %d replies.\n", p.clock, p.requestsSent)

	for id, client := range p.clients {
		_, err := client.Request(p.ctx, request)

		if err != nil {
			log.Printf("Something went wrong with %v\n", id)
		}
	}
}

func (p *peer) sendReplyToAllDefered() {

	p.incrementClock()

	if len(p.defered) != 0 {
		log.Printf("(L: %d) SEND: Sending reply to all deferred peers.\n", p.clock)
	} else {
		log.Printf("(L: %d) SEND: No deferred peers.\n", p.clock)
	}

	reply := &api.Reply{
		Clock: p.clock,
		Id:    p.id,
	}

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
		if p.state == RELEASED {
			log.Println("Requesting to work in critical section...")
			p.sendRequestToAll()
		} else {
			log.Println("Already working in critical section...")
		}
	}
}
