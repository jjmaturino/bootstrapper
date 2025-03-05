# Bootstrapper

A lightweight, extensible framework for Go services that can run on multiple platforms.

## Overview

Bootstrapper is designed to simplify the process of creating, configuring, and deploying services across different runtime environments. It provides clear separation between:

- **Service Logic**: Focus on your business logic, not infrastructure
- **Platform Runtimes**: Let the framework handle the environment-specific details
- **Service Type**: Support for HTTP services, queue processors, workers, and more

---

## Features
- **Multiple platform support**: Run your services on VMs, Docker, AWS Lambda, Kubernetes
- **Multiple service types**: HTTP services, Queue processors, Workers, gRPC services
- **Consistent interface**: Common initialization, configuration, and lifecycle
- **Extensible design**: Easy to add new platforms or service types
- **Logging**: Built-in integration with zap logger

## Current Features

- Virtual Machine (VM) runtime support
- HTTP service type with Gin integration
- Default middleware for logging and error handling
- Easy service initialization with dependency injection

## Future Extensibility

The framework is designed to be extended with:

- Additional platform types (Docker, Kubernetes, AWS Lambda, etc.)
- Additional service types (Queue processors, Workers, Scheduled tasks, etc.)
- Custom platform implementations through the `RegisterPlatform` function

---

## Installation

```bash
go get github.com/jjmaturino/bootstrapper
```

## Basic Usage

```go
package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jjmaturino/bootstrapper/launcher"
	"github.com/jjmaturino/bootstrapper/platform"
	"go.uber.org/zap"
)


func main() {
	// Initialize context
	ctx := context.Background()

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create Gin engine with default configuration
	engine:= gin.Default()

	// Create your service 
	service := NewMyHttpService() // Implement platform.ApiService

	// Start the service on VM platform, with dependencies injected
	launcher.Start(ctx, service, platform.VM, engine, logger)
}
```

## Creating a Service

To create a service, implement the `platform.ApiService` interface:

```go
package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jjmaturino/bootstrapper/launcher"
	"github.com/jjmaturino/bootstrapper/platform"
	"go.uber.org/zap"
)

// MyHTTPService implements the HTTPService interface
type MyHTTPService struct {
    logger *zap.Logger
}

func NewMyHTTPService() *MyHTTPService {
    return &MyHTTPService{}
}

// Initialize sets up the service
func (s *MyHTTPService) Initialize(ctx context.Context, deps ...interface{}) error {
    // Process dependencies
    for _, dep := range deps {
        if logger, ok := dep.(*zap.Logger); ok {
            s.logger = logger
        }
    }
    return nil
}

// Type returns the service type
func (s *MyHTTPService) Type() platform.ServiceType {
    return platform.HTTPService
}

// ConfigureRoutes sets up HTTP routes
func (s *MyHTTPService) ConfigureRoutes(ctx context.Context, engine platform.Engine) error {
    engine.Handle("GET", "/hello", func(c *gin.Context) {
        c.JSON(200, []byte(`{"message": "Hello, World!"}`))
    })
    
    return nil
}
```

## Extending with New Platforms

You can register custom platform implementations:

```go

package examplepkg

import (
    "context"
    "github.com/jjmaturino/bootstrapper/launcher"
    "github.com/jjmaturino/bootstrapper/platform"
)


// Define a custom Kubernetes service starter
type K8sServiceStarter struct {
	logger *zap.Logger
}

func (k *K8sServiceStarter) StartService(ctx context.Context, service platform.Service, deps ...interface{}) error {
	// Kubernetes-specific implementation
	return nil
}

var platformtype = "fly.io"

// Register it with the launcher
launcher.RegisterPlatform(ctx, platformtype, &K8sServiceStarter{logger})

// Now you can use it
launcher.Start(ctx, myService, platformtype, deps...)
```


## Architecture


``` 
┌─────────────────┐      ┌───────────────────┐
│     Service     │      │      Launcher     │
│  (Your Code)    │─────▶│ (Start Services)  │
└─────────────────┘      └─────────┬─────────┘
                                   │
                                   ▼
                        ┌─────────────────────┐
                        │  Platform Registry  │
                        └──────────┬──────────┘
                                   │
        ┌────────────────┬─────────┼─────────┬────────────────┐
        │                │         │         │                │
        ▼                ▼         ▼         ▼                ▼
┌─────────────────┐┌─────────────┐┌───────┐┌─────────┐┌─────────────────┐
│  VM Starter     ││ K8s Starter ││Docker ││ Lambda  ││ Custom Starter  │
└─────────────────┘└─────────────┘└───────┘└─────────┘└─────────────────┘
```
