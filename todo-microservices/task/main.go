package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pbProject "todo/proto/project"
	pbTask "todo/proto/task"
	db "todo/shared/db"

	"google.golang.org/grpc"
)

const port = ":50051"

func main() {
	projectConn, err := grpc.Dial(os.Getenv("PROJECT_SERVICE_ADDR"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial to project: %s", err)
	}

	// TODO: インターセプター
	dbConn := db.ConnectDB()
	srv := grpc.NewServer()
	// service登録
	pbTask.RegisterTaskServiceServer(srv, &TaskService{
		db:            dbConn,
		projectClient: pbProject.NewProjectServiceClient(projectConn),
	})

	// クライアントからのgRPC接続を待ち受ける
	go func() {
		listener, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to create listener: %s", err)
		}
		log.Println("start server on port", port)
		if err := srv.Serve(listener); err != nil {
			log.Println("failed to exit serve: ", err)
		}
	}()

	// グレースフルストップ
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM)
	// sigint チャネルに値が送信されるまで(=受信するまで)待機
	<-sigint
	// シグナルを受け取ったら graceful shutdown 開始
	log.Println("received a signal of graceful shutdown")
	stopped := make(chan struct{})
	go func() {
		// GracefulStop は、新しいgRPC接続とRPCの受け付けをやめて、処理中のすべてのRPCが終了するまでブロック
		srv.GracefulStop()
		close(stopped)
	}()
	// タイムアウト1min
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	select {
	case <-ctx.Done():
		// タイムアウト（1min経過）の場合Stopする
		srv.Stop()
	case <-stopped:
		cancel()
	}
	log.Println("completed graceful shutdown")
}
