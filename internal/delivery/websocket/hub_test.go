package websocket

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	fastws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
)

func TestShutdownStopsHubLoop(t *testing.T) {
	hub := NewHub()
	hub.Start()
	if err := hub.Shutdown(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-hub.runDone:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("hub loop did not stop after shutdown")
	}
}

func TestShutdownUnblocksRegistrationInProgress(t *testing.T) {
	hub := NewHub()
	hub.register = make(chan registerRequest, 1)
	server := startTestServerWithHub(t, hub)
	request := registerRequest{client: NewClient(hub, nil, "registering"), accepted: make(chan bool)}
	go func() { hub.register <- request }()
	waitForClientCount(t, hub, 1)

	_ = server.dial(t)
	deadline := time.Now().Add(time.Second)
	for len(hub.register) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := len(hub.register); got != 1 {
		t.Fatalf("queued production registration = %d, want 1", got)
	}
	if err := hub.Shutdown(); err != nil {
		t.Fatal(err)
	}
	waitForClientCount(t, hub, 0)
}

type testServer struct {
	hub     *Hub
	address string
	app     *fiber.App
	appDone chan error
	hubDone chan struct{}
}

func startTestServer(t *testing.T) *testServer {
	return startTestServerWithHub(t, NewHub())
}

func startTestServerWithHub(t *testing.T, hub *Hub) *testServer {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	RegisterRoutes(app, hub)
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	hub.Start()
	server := &testServer{hub: hub, address: listener.Addr().String(), app: app, appDone: make(chan error, 1), hubDone: hub.runDone}
	go func() { server.appDone <- app.Listener(listener) }()
	t.Cleanup(func() {
		if err := hub.Shutdown(); err != nil {
			t.Errorf("stop WebSocket hub: %v", err)
		}
		if err := app.ShutdownWithTimeout(time.Second); err != nil {
			t.Errorf("stop test HTTP server: %v", err)
		}
		select {
		case err := <-server.appDone:
			if err != nil {
				t.Errorf("test HTTP server: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("test HTTP server did not stop")
		}
	})
	return server
}

func (s *testServer) dial(t *testing.T) *fastws.Conn {
	t.Helper()
	url := "ws://" + s.address + "/ws/"
	conn, _, err := (&fastws.Dialer{HandshakeTimeout: time.Second}).Dial(url, http.Header{})
	if err != nil {
		t.Fatalf("connect to %s: %v", url, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func waitForClientCount(t *testing.T, hub *Hub, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := hub.ClientCount(); got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("client count = %d, want %d", hub.ClientCount(), want)
}

func TestRoutesBroadcastRawMessageToTwoClients(t *testing.T) {
	server := startTestServer(t)
	first, second := server.dial(t), server.dial(t)
	waitForClientCount(t, server.hub, 2)

	const message = "raw broadcast payload"
	if err := first.WriteMessage(fastws.TextMessage, []byte(message)); err != nil {
		t.Fatal(err)
	}
	for _, client := range []*fastws.Conn{first, second} {
		if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		kind, got, err := client.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if kind != fastws.TextMessage || string(got) != message {
			t.Fatalf("broadcast = (%d, %q), want raw text %q", kind, got, message)
		}
	}
}

func TestDisconnectRemovesClientWithoutBreakingBroadcast(t *testing.T) {
	server := startTestServer(t)
	first, second := server.dial(t), server.dial(t)
	waitForClientCount(t, server.hub, 2)
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	waitForClientCount(t, server.hub, 1)

	server.hub.Broadcast([]byte("still connected"))
	if err := second.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	_, got, err := second.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "still connected" {
		t.Fatalf("broadcast = %q", got)
	}
}

func TestDisconnectStopsPendingWriterBeforeShutdown(t *testing.T) {
	server := startTestServer(t)
	conn := server.dial(t)
	waitForClientCount(t, server.hub, 1)

	server.hub.mu.RLock()
	var client *Client
	for active := range server.hub.clients {
		client = active
		break
	}
	server.hub.mu.RUnlock()
	if client == nil {
		t.Fatal("registered client not found")
	}

	payload := bytes.Repeat([]byte("x"), 16<<20)
	for i := 0; i < 16; i++ {
		server.hub.Broadcast(payload)
	}
	deadline := time.Now().Add(time.Second)
	for len(client.send) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(client.send) == 0 {
		t.Fatal("large broadcasts did not back up behind the writer")
	}
	time.Sleep(50 * time.Millisecond)
	if len(client.send) == 0 {
		t.Fatal("writer drained every queued message before disconnect")
	}

	tcpConn, ok := conn.NetConn().(*net.TCPConn)
	if !ok {
		t.Fatalf("client connection type = %T, want *net.TCPConn", conn.NetConn())
	}
	if err := tcpConn.CloseWrite(); err != nil {
		t.Fatal(err)
	}
	waitForClientCount(t, server.hub, 0)
	if err := server.hub.Shutdown(); err != nil {
		t.Fatalf("shutdown after disconnect: %v", err)
	}
	select {
	case <-client.writeDone:
	default:
		t.Fatal("client writer remained alive after shutdown")
	}
}

func TestSlowClientIsEvictedWhenItsQueueIsFull(t *testing.T) {
	server := startTestServer(t)
	_ = server.dial(t)
	waitForClientCount(t, server.hub, 1)

	payload := bytes.Repeat([]byte("x"), 1<<20)
	broadcastDone := make(chan struct{})
	go func() {
		for i := 0; i < 512 && server.hub.ClientCount() != 0; i++ {
			server.hub.Broadcast(payload)
		}
		close(broadcastDone)
	}()
	select {
	case <-broadcastDone:
	case <-time.After(3 * time.Second):
		t.Fatal("broadcasts remained blocked while evicting the slow client")
	}
	if got := server.hub.ClientCount(); got != 0 {
		t.Fatalf("client count after slow-client eviction = %d, want 0", got)
	}
	if err := server.hub.Shutdown(); err != nil {
		t.Fatalf("stop WebSocket client pumps: %v", err)
	}
}

func TestShutdownDuringRegistrationAndBroadcastStopsAllWorkers(t *testing.T) {
	server := startTestServer(t)
	initial := server.dial(t)
	waitForClientCount(t, server.hub, 1)
	second := server.dial(t)
	waitForClientCount(t, server.hub, 2)

	server.hub.Broadcast([]byte("before shutdown"))
	for _, client := range []*fastws.Conn{initial, second} {
		if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		_, message, err := client.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if string(message) != "before shutdown" {
			t.Fatalf("broadcast = %q, want pre-shutdown message", message)
		}
	}

	const dialers, broadcasters = 16, 64
	start := make(chan struct{})
	dialStarted := make(chan struct{}, dialers)
	broadcastStarted := make(chan struct{}, broadcasters)
	heldDialReady := make(chan struct{}, 1)
	releaseHeldDial := make(chan struct{})
	releaseDialers := make(chan struct{})
	stopBroadcasts := make(chan struct{})
	var releaseHeldOnce, releaseDialersOnce, stopBroadcastsOnce sync.Once
	releaseHeld := func() { releaseHeldOnce.Do(func() { close(releaseHeldDial) }) }
	releaseDials := func() { releaseDialersOnce.Do(func() { close(releaseDialers) }) }
	stopBroadcasting := func() { stopBroadcastsOnce.Do(func() { close(stopBroadcasts) }) }
	t.Cleanup(releaseHeld)
	t.Cleanup(releaseDials)
	t.Cleanup(stopBroadcasting)
	var workers sync.WaitGroup
	workers.Add(dialers + broadcasters)
	for i := 0; i < dialers; i++ {
		dialer := &fastws.Dialer{HandshakeTimeout: 3 * time.Second}
		if i == dialers-1 {
			dialer.NetDialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
				conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
				if err != nil {
					return nil, err
				}
				heldDialReady <- struct{}{}
				select {
				case <-releaseHeldDial:
					return conn, nil
				case <-ctx.Done():
					_ = conn.Close()
					return nil, ctx.Err()
				}
			}
		}
		go func(dialer *fastws.Dialer) {
			defer workers.Done()
			<-start
			dialStarted <- struct{}{}
			conn, _, err := dialer.Dial("ws://"+server.address+"/ws/", nil)
			if err == nil {
				<-releaseDialers
				_ = conn.Close()
			}
		}(dialer)
	}
	for i := 0; i < broadcasters; i++ {
		go func() {
			defer workers.Done()
			<-start
			server.hub.Broadcast([]byte("during shutdown"))
			broadcastStarted <- struct{}{}
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-stopBroadcasts:
					return
				case <-ticker.C:
					server.hub.Broadcast([]byte("during shutdown"))
				}
			}
		}()
	}
	close(start)
	for i := 0; i < dialers; i++ {
		<-dialStarted
	}
	select {
	case <-heldDialReady:
	case <-time.After(time.Second):
		t.Fatal("pending WebSocket dial did not establish a TCP connection")
	}
	for i := 0; i < broadcasters; i++ {
		<-broadcastStarted
	}
	deadline := time.Now().Add(time.Second)
	for server.hub.ClientCount() < 3 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := server.hub.ClientCount(); got < 3 {
		t.Fatalf("client count during concurrent registration = %d, want at least 3", got)
	}
	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- server.hub.Shutdown() }()
	select {
	case <-server.hub.done:
	case <-time.After(time.Second):
		t.Fatal("hub shutdown did not begin")
	}
	server.hub.Broadcast([]byte("while stopping"))
	releaseHeld()
	if err := <-shutdownDone; err != nil {
		t.Fatal(err)
	}
	releaseDials()
	stopBroadcasting()
	_ = initial.Close()

	select {
	case <-server.hubDone:
	case <-time.After(time.Second):
		t.Fatal("hub loop remained alive after shutdown")
	}
	workersDone := make(chan struct{})
	go func() { workers.Wait(); close(workersDone) }()
	select {
	case <-workersDone:
	case <-time.After(2 * time.Second):
		t.Fatal("registration or broadcast remained blocked during shutdown")
	}
	waitForClientCount(t, server.hub, 0)
}
