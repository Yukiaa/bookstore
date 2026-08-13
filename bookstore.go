package main

import (
	"bookstore/pb"
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultCursor   = "0"
	defaultPageSize = 2
)

type bookstore struct {
	pb.UnimplementedBookstoreServer
	store *Store
}

func NewBookstore(store *Store) *bookstore {
	return &bookstore{store: store}
}

func (b *bookstore) CreateShelf(ctx context.Context, req *pb.CreateShelfRequest) (*pb.Shelf, error) {
	shelf := &Shelf{Theme: req.GetShelf().GetTheme(), Size: req.GetShelf().GetSize()}
	created, err := b.store.CreateShelf(ctx, shelf)
	if err != nil {
		return nil, err
	}
	return &pb.Shelf{
		Id:    created.ID,
		Theme: created.Theme,
		Size:  created.Size,
	}, nil
}

func (b *bookstore) ListShelves(ctx context.Context, req *emptypb.Empty) (*pb.ListShelvesResponse, error) {
	shelves, err := b.store.ListShelves(ctx)
	if err != nil {
		return nil, err
	}
	var result []*pb.Shelf
	for _, s := range shelves {
		result = append(result, &pb.Shelf{
			Id:    s.ID,
			Theme: s.Theme,
			Size:  s.Size,
		})
	}
	return &pb.ListShelvesResponse{Shelves: result}, nil
}

func (b *bookstore) GetShelf(ctx context.Context, req *pb.GetShelfRequest) (*pb.Shelf, error) {
	shelf, err := b.store.GetShelf(ctx, req.GetShelf())
	if err != nil {
		return nil, err
	}
	return &pb.Shelf{
		Id:    shelf.ID,
		Theme: shelf.Theme,
		Size:  shelf.Size,
	}, nil
}

func (b *bookstore) DeleteShelf(ctx context.Context, req *pb.DeleteShelfRequest) (*emptypb.Empty, error) {
	if err := b.store.DeleteShelf(ctx, req.GetShelf()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (b *bookstore) ListBooks(ctx context.Context, in *pb.ListBooksRequest) (*pb.ListBooksReponse, error) {
	if in.Shelf <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invaild shelf")
	}
	var (
		cursor   = defaultCursor
		pageSize = defaultPageSize
	)
	// if in.PageToken==""{
	// 	//没有分页默认第一页
	// }else {
	if len(in.PageToken) > 0 {
		pageInfo := Token(in.PageToken).Decode()
		if pageInfo.IsInVaild() {
			return nil, status.Error(codes.InvalidArgument, "invaild pagetoken")
		}
		//有分页就先解析分页
		cursor = pageInfo.NextID
		pageSize = int(pageInfo.PageSize)
	}

	//查询数据库
	booklist, err := b.store.GetBookListByShelfID(ctx, in.Shelf, cursor, pageSize+1)
	if err != nil {
		fmt.Printf("GetBookListByShelfID fail %v", err)
		return nil, status.Error(codes.Internal, "query fail")
	}
	var (
		hasNextPage   bool
		nextPageToken string
		realSize      int = len(booklist)
	)
	if len(booklist) > pageSize {
		hasNextPage = true
		realSize = pageSize

	}
	//封装返回的数据
	res := make([]*pb.Book, 0, len(booklist))
	//
	for i := 0; i < realSize; i++ {
		res = append(res, &pb.Book{
			Id:     booklist[i].ID,
			Author: booklist[i].Author,
		})
	}
	if hasNextPage {
		nextPageInfo := Page{
			NextID:        strconv.FormatInt(res[realSize-1].Id, 10),
			NextTimeAtUTC: time.Now().Unix(),
			PageSize:      int64(pageSize),
		}
		nextPageToken = string(nextPageInfo.Encode())
	}
	return &pb.ListBooksReponse{Book: res, NextPageToken: nextPageToken}, nil
}
