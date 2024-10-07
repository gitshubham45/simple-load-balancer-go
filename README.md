# Simple Load Balancer in Go

## Overview
This Go application implements a basic load balancer that distributes incoming HTTP requests across multiple backend servers using a round-robin algorithm.

## Features
1. **Round-Robin Request Distribution**: Balances requests evenly across backend servers in a round-robin fashion.
2. **Health Checks**: Verifies server availability before forwarding requests to ensure only healthy servers are used.
3. **Reverse Proxy**: Transparently forwards incoming requests to backend servers using `httputil.ReverseProxy`.
4. **Configurable Backends**: Add or remove backend servers by modifying the server list in the code.

## How to Run
1. Start your backend servers.
2. Update the server addresses in the load balancer code.
3. Run the load balancer:
   ```bash
   go run main.go
