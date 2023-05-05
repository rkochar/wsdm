package main

import (
	"context"
	"go-micro.dev/v4"
)

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

type Helloworld struct{}

func (h *Helloworld) Greeting(ctx context.Context, req *Request, rsp *Response) error {
	rsp.Message = "Hello " + req.Name
	return nil
}

func main() {
	// create a new service
	service := micro.NewService(
		//micro.Handle(":8080"),
		micro.Name("helloworld"),
		micro.Address("localhost:8080"),
		micro.Handle(new(Helloworld)),
	)

	// initialise flags
	service.Init()

	// start the service
	service.Run()
}
