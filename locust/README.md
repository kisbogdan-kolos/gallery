# Load testing

## Locustfile

The locustfile does the following actions:

- register a user
- log in with the user
- get all images, and view 20 randomly selected ones
- upload a new image
- delete uploaded images

The script cleans up all the uploaded images when exiting.

## Building and running

### Locally

When running locally, you only need to install `locust` on your computer, and run 

```bash
locust
```

Then open a web browser on the given URL, and start the tests.

### Kubernetes/OpenShift

First, you need to build a Docker image containing the locustfile and the test images

```bash
docker build -t kisbogdan/gallery-locust:latest .
docker push kisbogdan/gallery-locust:latest
```

Then, you need to deploy it using the given `locust.yaml`

```bash
kubectl apply -f locust.yaml
kubectl port-forward services/locust 8089:8089
```

Or, if running OpenShift

```bash
oc apply -f locust.yaml
oc port-forward services/locust 8089:8089
```
