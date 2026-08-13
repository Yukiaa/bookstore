package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "bookstore/bookstoreclient/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	conn, err := grpc.NewClient(
		"localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBookstoreClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch os.Args[1] {
	case "list":
		listShelves(ctx, client)
	case "create":
		if len(os.Args) < 4 {
			fmt.Println("Usage: create <theme> <size>")
			os.Exit(1)
		}
		createShelf(ctx, client, os.Args[2], os.Args[3])
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("Usage: get <id>")
			os.Exit(1)
		}
		getShelf(ctx, client, os.Args[2])
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: delete <id>")
			os.Exit(1)
		}
		deleteShelf(ctx, client, os.Args[2])
	case "list-books":
		if len(os.Args) < 3 {
			fmt.Println("Usage: list-books <shelf_id>")
			os.Exit(1)
		}
		listBooks(ctx, client, os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Bookstore gRPC Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  list                         List all shelves")
	fmt.Println("  create <theme> <size>        Create a new shelf")
	fmt.Println("  get <id>                     Get a shelf by ID")
	fmt.Println("  delete <id>                  Delete a shelf by ID")
	fmt.Println("  list-books <shelf_id>        List books on a shelf")
}

func listShelves(ctx context.Context, client pb.BookstoreClient) {
	resp, err := client.ListShelves(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("ListShelves failed: %v", err)
	}
	if len(resp.GetShelves()) == 0 {
		fmt.Println("No shelves found.")
		return
	}
	fmt.Println("Shelves:")
	for _, s := range resp.GetShelves() {
		fmt.Printf("  ID: %d, Theme: %s, Size: %d\n", s.GetId(), s.GetTheme(), s.GetSize())
	}
}

func createShelf(ctx context.Context, client pb.BookstoreClient, theme, sizeStr string) {
	var size int64
	if _, err := fmt.Sscanf(sizeStr, "%d", &size); err != nil {
		log.Fatalf("invalid size: %v", err)
	}
	resp, err := client.CreateShelf(ctx, &pb.CreateShelfRequest{
		Shelf: &pb.Shelf{
			Theme: theme,
			Size:  size,
		},
	})
	if err != nil {
		log.Fatalf("CreateShelf failed: %v", err)
	}
	fmt.Printf("Shelf created: ID=%d, Theme=%s, Size=%d\n", resp.GetId(), resp.GetTheme(), resp.GetSize())
}

func getShelf(ctx context.Context, client pb.BookstoreClient, idStr string) {
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		log.Fatalf("invalid id: %v", err)
	}
	resp, err := client.GetShelf(ctx, &pb.GetShelfRequest{Shelf: id})
	if err != nil {
		log.Fatalf("GetShelf failed: %v", err)
	}
	fmt.Printf("Shelf: ID=%d, Theme=%s, Size=%d\n", resp.GetId(), resp.GetTheme(), resp.GetSize())
}

func deleteShelf(ctx context.Context, client pb.BookstoreClient, idStr string) {
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		log.Fatalf("invalid id: %v", err)
	}
	_, err := client.DeleteShelf(ctx, &pb.DeleteShelfRequest{Shelf: id})
	if err != nil {
		log.Fatalf("DeleteShelf failed: %v", err)
	}
	fmt.Printf("Shelf %d deleted successfully.\n", id)
}

func listBooks(ctx context.Context, client pb.BookstoreClient, shelfIDStr string) {
	var shelfID int64
	if _, err := fmt.Sscanf(shelfIDStr, "%d", &shelfID); err != nil {
		log.Fatalf("invalid shelf id: %v", err)
	}
	resp, err := client.ListBooks(ctx, &pb.ListBooksRequest{Shelf: shelfID})
	if err != nil {
		log.Fatalf("ListBooks failed: %v", err)
	}
	if len(resp.GetBook()) == 0 {
		fmt.Printf("No books found on shelf %d.\n", shelfID)
		return
	}
	fmt.Printf("Books on shelf %d:\n", shelfID)
	for _, b := range resp.GetBook() {
		fmt.Printf("  ID: %d, Author: %s\n", b.GetId(), b.GetAuthor())
	}
}
