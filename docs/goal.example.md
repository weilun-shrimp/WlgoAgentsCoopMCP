# Goal Example

This is a testing case. I gave this content to the Manager agent, and it coordinated all agents to build everything.

```
Dockerfile
ubuntu 24.04
golang 1.25
Fiber http server
.env and .env.example
- LISTEN_PORT - It for http server
- APP_NAME
A README.md to show how to use docker to run this service.
- How to build image
- How to run container
I would like do flexible when run docker.
It means I don't like EXPOSE port in Dockerfile.
I like use -p attribute when docker run
eg. -w, -v, -p.....
Therefore You only need to install basic requirements in Dockerfile.

Service includings
- Graceful shutdown.
- Post manager
    - A slice to store posts
    - CRUD posts


Use https://github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr to manage servies
Folder struct
/internal
    /service1
    /service2
eg.
/internal
    /kernel
        quit_signal.go (for graceful shutdown)
        http_server.go (for fiber) when go wrong also need to grance shut down.
    /post
        ..... (impleemnt Post service here)

API Includings
- handler (Ping) return "Pong"
- CRUD post handlers to post service


```
