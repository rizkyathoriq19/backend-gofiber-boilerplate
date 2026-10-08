package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"boilerplate-be/internal/config"
	"boilerplate-be/internal/database"

	"github.com/gofiber/fiber/v2"
)

func TestServeReturnsBindError(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := serve(ctx, app, listener.Addr().String()); err == nil {
		t.Fatal("expected occupied address to fail")
	} else {
		var networkError *net.OpError
		if !errors.As(err, &networkError) {
			t.Fatalf("expected bind error, got %v", err)
		}
	}
}

func TestServeCancellationClosesListener(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })
	listening := make(chan string, 1)
	app.Hooks().OnListen(func(data fiber.ListenData) error {
		listening <- net.JoinHostPort(data.Host, data.Port)
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Cleanup(func() { _ = app.ShutdownWithTimeout(time.Second) })
	done := make(chan error, 1)
	go func() { done <- serve(ctx, app, "127.0.0.1:0") }()
	var address string
	select {
	case address = <-listening:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not listen")
	}
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + address + "/ping")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	client.CloseIdleConnections()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("ping status = %d", response.StatusCode)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation did not stop server")
	}
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		t.Fatalf("listener was not released: %v", err)
	}
	listener.Close()
}

func TestServeCancellationDuringStartup(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	listening, resume := make(chan struct{}), make(chan struct{})
	app.Hooks().OnListen(func(fiber.ListenData) error {
		close(listening)
		<-resume
		return nil
	})
	shutdown := make(chan struct{}, 1)
	app.Hooks().OnShutdown(func() error {
		select {
		case shutdown <- struct{}{}:
		default:
		}
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- serve(ctx, app, "127.0.0.1:0") }()
	select {
	case <-listening:
	case <-time.After(2 * time.Second):
		close(resume)
		t.Fatal("server did not start")
	}
	cancel()
	select {
	case <-shutdown:
	case <-time.After(2 * time.Second):
		close(resume)
		t.Fatal("shutdown did not begin")
	}
	close(resume)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		_ = app.ShutdownWithTimeout(time.Second)
		t.Fatal("startup cancellation leaked listener")
	}
}

func TestServeDrainsInflightRequest(t *testing.T) {
	t.Run("graceful", func(t *testing.T) {
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		listening := make(chan string, 1)
		app.Hooks().OnListen(func(data fiber.ListenData) error {
			listening <- net.JoinHostPort(data.Host, data.Port)
			return nil
		})
		started, release := make(chan struct{}), make(chan struct{})
		app.Get("/slow", func(c *fiber.Ctx) error {
			close(started)
			<-release
			return c.SendStatus(fiber.StatusOK)
		})
		defer close(release)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- serve(ctx, app, "127.0.0.1:0") }()
		var address string
		select {
		case address = <-listening:
		case <-time.After(2 * time.Second):
			t.Fatal("server did not listen")
		}
		requestDone := make(chan error, 1)
		go func() {
			client := &http.Client{Timeout: 15 * time.Second}
			defer client.CloseIdleConnections()
			response, err := client.Get("http://" + address + "/slow")
			if err == nil {
				response.Body.Close()
			}
			requestDone <- err
		}()
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("request did not reach handler")
		}
		cancel()
		select {
		case err := <-done:
			t.Fatalf("shutdown returned before request drained: %v", err)
		case <-time.After(50 * time.Millisecond):
		}
		release <- struct{}{}
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("shutdown did not finish after the request drained")
		}
		if err := <-requestDone; err != nil {
			t.Fatalf("in-flight request interrupted: %v", err)
		}
	})
}

// Run against dedicated, disposable PostgreSQL and Redis instances. No schema is needed.
func TestRunClosesResourcesOnExit(t *testing.T) {
	postgresPort, redisPort := os.Getenv("TEST_POSTGRES_PORT"), os.Getenv("TEST_REDIS_PORT")
	if postgresPort == "" || redisPort == "" {
		t.Skip("set TEST_POSTGRES_PORT and TEST_REDIS_PORT for disposable loopback services (postgres/password)")
	}
	t.Chdir(t.TempDir()) // run must not load the repository's .env file.
	for key, value := range map[string]string{
		"DB_HOST": "127.0.0.1", "DB_PORT": postgresPort,
		"DB_USER": "postgres", "DB_PASSWORD": "password", "DB_NAME": "postgres", "DB_SSL_MODE": "disable",
		"REDIS_HOST": "127.0.0.1", "REDIS_PORT": redisPort, "REDIS_PASSWORD": "", "REDIS_DB": "0",
		"APP_ENV": "production", "APP_PREFORK": "false",
	} {
		t.Setenv(key, value)
	}
	cfg := config.New()
	db, err := database.New(cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		t.Fatal(err)
	}
	defer redisClient.Close()
	counts := func() (int, int) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		var connections int
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND pid <> pg_backend_pid()").Scan(&connections); err != nil {
			t.Fatal(err)
		}
		clients, err := redisClient.ClientList(ctx).Result()
		if err != nil {
			t.Fatal(err)
		}
		return connections, len(strings.FieldsFunc(clients, func(r rune) bool { return r == '\n' }))
	}
	baseDB, baseRedis := counts()
	for _, failure := range []string{"redis", "listen", "shutdown"} {
		t.Run(failure, func(t *testing.T) {
			if failure == "redis" {
				t.Setenv("REDIS_PORT", "invalid")
			} else {
				listener, err := net.Listen("tcp4", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				_, port, _ := net.SplitHostPort(listener.Addr().String())
				t.Setenv("APP_PORT", port)
				if failure == "shutdown" {
					listener.Close()
				} else {
					defer listener.Close()
				}
			}
			if failure == "shutdown" {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				done := make(chan error, 1)
				go func() { done <- run(ctx) }()
				client := &http.Client{Timeout: time.Second}
				defer client.CloseIdleConnections()
				url := "http://127.0.0.1:" + os.Getenv("APP_PORT") + "/ping"
				deadline := time.Now().Add(3 * time.Second)
				for {
					resp, err := client.Get(url)
					if err == nil {
						resp.Body.Close()
						if resp.StatusCode != http.StatusOK {
							t.Fatalf("production ping status: %d", resp.StatusCode)
						}
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("production server did not listen")
					}
					time.Sleep(10 * time.Millisecond)
				}
				activeDB, activeRedis := counts()
				if activeDB <= baseDB || activeRedis < baseRedis+2 {
					t.Fatal("production resources were not acquired")
				}
				client.CloseIdleConnections()
				cancel()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("production shutdown did not finish")
				}
				listener, err := net.Listen("tcp4", "127.0.0.1:"+os.Getenv("APP_PORT"))
				if err != nil {
					t.Fatalf("production listener was not released: %v", err)
				}
				listener.Close()
			} else if err := run(context.Background()); err == nil || !strings.Contains(strings.ToLower(err.Error()), failure) {
				t.Fatalf("expected %s failure, got %v", failure, err)
			}
			deadline := time.Now().Add(2 * time.Second)
			for {
				gotDB, gotRedis := counts()
				if gotDB == baseDB && gotRedis == baseRedis {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("leaked connections: postgres %d -> %d, redis %d -> %d", baseDB, gotDB, baseRedis, gotRedis)
				}
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}
