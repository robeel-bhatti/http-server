# Custom HTTP Server

This project is a basic HTTP server written in Go (Golang). 
Its purpose is to deepen understanding of how HTTP servers work under the hood—specifically what happens before 
an HTTP request reaches the application-level handlers that developers typically write.

# Why This Project

This project taught me what happens behind the scenes with web servers that host and support REST APIs.

I gained a firm understanding of:

- The TCP/IP model and how network communication works at a fundamental level

- The HTTP specification and how requests/responses must be formatted to comply with the spec so machines can communicate using the HTTP protocol (I realized that building CRUD APIs only teaches you what HTTP is on a surface level)

- Golang itself and why it's well-suited for systems programming:
    - Simplifies concurrency handling and thread management compared to other languages
    - Clean, non-verbose syntax with straightforward error handling (no abstractions)
    - Pointers teach you how objects are managed in memory, enabling you to build performance-intensive applications through deliberate memory management
    - Performance: Go compiles directly to machine code, so the CPU executes raw instructions without any interpreter or VM overhead.

# Running Locally

Run `run.sh`