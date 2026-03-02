# Simple gallery app

## Deploying

Currently, the build can only be done locally, and there is no CI/CD. This will change in the future.

```bash
cd src
docker build -t kisbogdan/gallery:latest .
docker push kisbogdan/gallery:latest
```

> Warning: Currently, the env variables are burned into the `deploy.yaml`, so the deployment is **not** rpoduction-ready.

### Kubernetes

Use the `deploy.yaml` to deploy to Kubernetes.

```bash
kubectl apply -f deploy.yaml
```

### OpenShift (BME Cloud)

Modify the `deploy.yaml`.

1. remove the `PersistetVolume` section
1. modify the `PersistentVolumeClaim` `storageClassName` from `manual` to `rook-cephfs`

Deploy using the modified config using OC command line tool.

```bash
oc apply -f deploy.yaml
```

Or deploy by pasting the content of the file into the web console _from YAML_ section.

Create a new route to expose the service:

```yaml
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: gallery-svc
  namespace: lab2
spec:
  to:
    name: gallery
    weight: 100
    kind: Service
  host: ''
  path: ''
  tls:
    insecureEdgeTerminationPolicy: ''
    termination: edge
  port:
    targetPort: 8080
  alternateBackends: []
```

Now, your app should be accessible from _https://gallery.svc.apps.okf.fured.cloud.bme.hu/_.

### API
