# wait-for

A command line tool to wait for network healthy state to start another command, and then wait for network unhealthy state to stop it.

```
  -wait-start url
    	wait for url to become healthy to start
  -wait-start-interval duration
    	interval for wait-start (default 1s)
  -wait-start-retry uint
    	retry count for wait-start (default 10)
  -wait-start-timeout duration
    	timeout for wait-start (default 1s)
  -wait-stop url
    	wait for url to become unhealthy to stop
  -wait-stop-interval duration
    	interval for wait-stop (default 1s)
  -wait-stop-retry uint
    	retry count for wait-stop (default 10)
  -wait-stop-timeout duration
    	timeout for wait-stop (default 1s)
```

```bash
wait-for --wait-start=tcp://:8000,tcp://:8001 --wait-stop=tcp://:8002,tcp://:8003 -- sleep 60
```

## [Docker Pull Command](https://hub.docker.com/r/taylorchu/wait-for)

```
docker pull taylorchu/wait-for:latest
```

## Why?

### Kubernetes

Kubernetes supports multiple containers in a pod, but there is no current feature to manage dependency ordering, so all the containers (other than init containers) start at the same time. This can cause a number of issues with certain configurations:

1. [Kubernetes jobs](https://github.com/kubernetes/kubernetes/issues/25908) run until all containers have exited. If a sidecar container is supporting a primary container, the sidecar needs to be gracefully terminated after the primary container has exited, before the job will end.
2. [Sidecar proxies](https://github.com/GoogleCloudPlatform/cloud-sql-proxy/issues/128) (e.g. Istio, CloudSQL Proxy) are often designed to handle network traffic to and from a pod's primary container. But if the primary container tries to make egress call or recieve ingress calls before the sidecar proxy is up and ready, those calls may fail.

The k8s enhancement to address this is [sidecar container](https://github.com/kubernetes/enhancements/issues/753), but [that KEP will not be progressing](https://github.com/kubernetes/enhancements/issues/753#issuecomment-713471597). Later, another work-in-progress k8s enhancement about [keystone container](https://github.com/kubernetes/enhancements/issues/2872) appears, but it is [unclear that the KEP design is solid](https://github.com/kubernetes/enhancements/pull/2869#issuecomment-1270508226).

#### Install

1. Add `restartPolicy: Never` because we want to manage restart cycle ourselves.
2. Add volume and init container to copy `wait-for` binary that will be used in all containers in this pod:

```
volumes:
  - name: wait-for
    emptyDir: {}
initContainers:
  - name: wait-for
    image: taylorchu/wait-for:latest
    imagePullPolicy: IfNotPresent
    command: ['cp', '/usr/bin/wait-for', '/wait-for/wait-for']
    volumeMounts:
      - name: wait-for
        mountPath: /wait-for
```

3. For each container in this pod, add volume mount and command/args:

```
volumeMounts:
  - name: wait-for
    mountPath: /wait-for
command: ['/wait-for/wait-for', '--wait-start=tcp://some_address', '--wait-stop=tcp://some_address', '--']
args: ['some_command']
```

> Replace `some_address` and `some_command` with real values.

## Other ideas

- [jwilder/dockerize](https://github.com/jwilder/dockerize)
- [karlkfi/kubexit](https://github.com/karlkfi/kubexit)
- [nrmitchi/k8s-controller-sidecars](https://github.com/nrmitchi/k8s-controller-sidecars)
