package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"bookstore/pb"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//rpc http 不同端口
// func main() {
// 	dsn := "root:123456@tcp(127.0.0.1:3306)/bookstore?charset=utf8mb4&parseTime=True&loc=Local"
// 	db, err := NewDB(dsn)
// 	if err != nil {
// 		panic(err)
// 	}
// 	sqlDB, err := db.DB()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer sqlDB.Close()

// 	store := NewStore(db)
// 	b := NewBookstore(store)

// 	lis, err := net.Listen("tcp", ":8972")
// 	if err != nil {
// 		fmt.Printf("fail to create lis")
// 	}
// 	s := grpc.NewServer()
// 	pb.RegisterBookstoreServer(s, b)
// 	go func() {
// 		fmt.Println(s.Serve(lis))
// 	}()

// 	conn, err := grpc.DialContext(
// 		context.Background(),
// 		"127.0.0.1:8972",
// 		grpc.WithBlock(),
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 	)
// 	if err != nil {
// 		log.Fatalln("Failed to dial server:", err)
// 	}
// 	gwmux := runtime.NewServeMux()
// 	// 注册Greeter
// 	err = pb.RegisterBookstoreHandler(context.Background(), gwmux, conn)
// 	if err != nil {
// 		log.Fatalln("Failed to register gateway:", err)
// 	}
// 	//定义HTTP server配置
// 	gwServer := &http.Server{
// 		Addr:    ":8080",
// 		Handler: gwmux,
// 	}
// 	fmt.Println("grpc gateway qidong")
// 	gwServer.ListenAndServe()
// }

// rpc http 相同端口
func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/bookstore?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := NewDB(dsn)
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	store := NewStore(db)
	b := NewBookstore(store)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("fail to create lis")
	}
	s := grpc.NewServer()
	pb.RegisterBookstoreServer(s, b)

	gwmux := runtime.NewServeMux()
	dops := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = pb.RegisterBookstoreHandlerFromEndpoint(context.Background(), gwmux, "127.0.0.1:8080", dops)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	gwServer := &http.Server{
		Addr:    ":8080",
		Handler: grpcHandlerFunc(s, mux),
	}
	fmt.Println("server on :8080")
	gwServer.Serve(lis)

}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
