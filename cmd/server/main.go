package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fuzail/08-todosvc/internal/grpc"
	"github.com/fuzail/08-todosvc/internal/rest"
	"github.com/fuzail/08-todosvc/internal/todo"
	"github.com/fuzail/08-todosvc/pkg/db"
	pb "github.com/fuzail/08-todosvc/proto"
	grpcObj "google.golang.org/grpc"
)

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}

func main() {
	var migrateOnly bool
	flag.BoolVar(&migrateOnly, "migrate-only", false, "run migrations and exit")
	flag.Parse()

	// read env with sensible defaults
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// create DB
	dbConn, err := db.NewGormDBFromEnv()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	// run AutoMigrate (recommended for dev)
	if err := dbConn.AutoMigrate(&todo.Task{}); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	log.Println("migrations applied (AutoMigrate)")

	if migrateOnly {
		log.Println("migrate-only flag set; exiting")
		return
	}

	// wire repository and service
	repo := todo.NewGormRepository(dbConn)
	service := todo.NewService(repo)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpcObj.NewServer()
	pb.RegisterTodoServiceServer(grpcServer, grpc.NewHandler(service))

	// Start HTTP server (REST wrapper calling the service directly)
	mux := http.NewServeMux()
	rest.RegisterHandlers(mux, service)
	httpSrv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}

	// run servers concurrently
	serverErrCh := make(chan error, 2)
	go func() {
		log.Printf("gRPC listening on %s", lis.Addr())
		serverErrCh <- grpcServer.Serve(lis)
	}()

	go func() {
		log.Printf("HTTP server listening on %s", httpSrv.Addr)
		serverErrCh <- httpSrv.ListenAndServe()
	}()

	// graceful shutdown on signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case <-stop:
		log.Println("shutdown signal received")
	case err := <-serverErrCh:
		log.Printf("server error: %v", err)
	}

	// begin graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// stop gRPC (no graceful built-in; use Stop() to force immediate or GracefulStop to wait)
	go func() {
		grpcServer.GracefulStop()
	}()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	// close DB
	sqlDB, _ := dbConn.DB()
	_ = sqlDB.Close()

	log.Println("server stopped")
}
