# Getting Started

## Local Cluster (`kind`)

Install `kind` (if you don't have it already)

```shell
brew install kind
kind --version
 # -> kind version x.x.x
```

Start or create a local cluster

```shell
kind create cluster
```

Switch `kubectl` context to use your `kind` cluster

```shell
kubectl config set-context kind-kind
```

## Makefile

Test an initial build by simply running:

```shell
make
```

Test generating the manifest via:

```shell
make manifests
```

Install CRDs to the cluster (ensure you have `kind` running and your `kubectl` context is configured correctly - pointing at `kind`)

```shell
make install
```

## Debugging with GoLand

Set your breakpoint(s) (likely in `controllers/deploymentrule_controller.go`)

### Start the debugging session

From the `main.go` file

- Either use the GUI (right-click on the play button next to `func main()` and click Debug)
- Or, cool kids on Mac can press `CTRL + SHIFT + D` with the `main.go` file open.

After the first run just press `CTRL + D` from any file

### Apply a test `DeploymentRule`

```shell
kubectl apply -f config/samples/core_v1alpha1_deploymentrule.yaml
```

### Apply a test `Deployment`

```shell
kubectl apply -f hack/test-deploy.yml
```

If you want to apply the same deployment again you'll need to delete it first

```shell
kubectl delete -f hack/test-deploy.yml
```
