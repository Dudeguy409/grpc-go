/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	port       = flag.Int("port", 50051, "The server port")
	unreliable = flag.Bool("unreliable", false, "toggles serving status every 5 seconds")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	healthServer := health.NewServer()

	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(pb.Greeter_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	if *unreliable {
		go toggleServingStatus(healthServer)
	}

	pb.RegisterGreeterServer(s, &server{})
	healthpb.RegisterHealthServer(s, healthServer)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func toggleServingStatus(healthServer *health.Server) {
	for {
		status := healthpb.HealthCheckResponse_SERVING
		// Check if user Service is valid
		if time.Now().Second()%2 == 0 {
			status = healthpb.HealthCheckResponse_NOT_SERVING
		}

		healthServer.SetServingStatus(pb.Greeter_ServiceDesc.ServiceName, status)
		healthServer.SetServingStatus("", status)

		time.Sleep(5 * time.Second)
	}
}
