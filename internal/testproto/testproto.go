package testproto

//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative testproto.proto
