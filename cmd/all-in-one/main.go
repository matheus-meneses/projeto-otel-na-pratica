// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/infra/telemetry"
	"log"
	"net"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/app"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

func main() {

	tp, err := telemetry.InitTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.(*sdktrace.TracerProvider).Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	mp, err := telemetry.InitMeter()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mp.(*sdkmetric.MeterProvider).Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	configFlag := flag.String("config", "", "path to the config file")
	flag.Parse()

	c, _ := config.LoadConfig(*configFlag)

	mux := http.NewServeMux()

	// starts the gRPC server
	lis, _ := net.Listen("tcp", c.Server.Endpoint.GRPC)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	{
		a := app.NewUser(&c.Users)
		a.RegisterRoutes(mux)
	}

	{
		a := app.NewPlan(&c.Plans)
		a.RegisterRoutes(mux, grpcServer)
	}

	{
		a, err := app.NewPayment(&c.Payments)
		if err != nil {
			panic(err)
		}
		a.RegisterRoutes(mux)
		defer func() {
			_ = a.Shutdown()
		}()
	}

	{
		a := app.NewSubscription(&c.Subscriptions)
		a.RegisterRoutes(mux)
	}

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	_ = http.ListenAndServe(c.Server.Endpoint.HTTP, mux)
}
