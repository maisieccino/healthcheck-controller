# HealthCheck Controller

The HealthCheck Controller manages `HealthCheck` resources, which can be used
to monitor service health, cluster health, and more.

## Design

See [design.md](./design.md) for a high level design document.

## Installation

The easiest way of getting the controller running is by using a Docker image.

We currently aren't building public Docker images, but you can use the
provided Docker file to create one for yourself.

```bash
$ docker build -t hc-controller:local .
# [...]
```

If you are using Docker for Mac, you can run this image as a pod inside your
cluster without any extra steps:

```bash
$ kubectl run hc-controller \
        --generator=run-pod/v1 \
        --image=hc-controller:local
```

Otherwise, follow steps for your distribution to run a Docker image in your cluster.

You can also just run the program from your local machine, provided you have a kube config file:

```bash
$ go run cmd/hc-controller/main.go -kubeconfig=$HOME/.kube/config
I1117 15:55:19.416790   87637 controller.go:92] Setting up event handlers
I1117 15:55:19.416916   87637 controller.go:139] Starting HealthCheck controller
I1117 15:55:19.416920   87637 controller.go:141] Waiting for caches to sync
I1117 15:55:19.521876   87637 controller.go:146] Starting workers
I1117 15:55:19.521897   87637 controller.go:152] Started workers
```

## Development

We recommend using a tool like [Okteto](https://okteto.com) for easy local
development. Run `okteto init`, accept the default values, and then run
`okteto up` to bring up a shell inside the cluster. Then you can run `go run
cmd/hc-controller/main.go` to run the controller in-cluster.

Files are synced automatically between your host laptop and the Okteto pod.

## Testing

Run test suite with `go test ./...`. We use `testify` to define tests.
