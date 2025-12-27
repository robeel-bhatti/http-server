# Custom HTTP Server

This project is a basic HTTP server written in Go (Golang). 
Its purpose is to deepen understanding of how HTTP servers work under the hood—specifically what happens before 
an HTTP request reaches the application-level handlers that developers typically write.

Rather than relying on high-level frameworks, this server focuses on the lower-level mechanics involved 
in handling requests, providing insight into how Go manages networking and concurrency.

# Why This Project

Most production servers hide these details behind frameworks and abstractions. 
This project intentionally peels back those layers to help build a stronger mental model of how:

- TCP connections are established and managed
- HTTP requests are received, parsed, and routed
- Servers handle multiple clients concurrently