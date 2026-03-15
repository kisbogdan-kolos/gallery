# Simple gallery app

## Deploying

The project is equipped with a GitHub Actions workflow that automatically builds and pushes the Docker image to the registry upon changes to the `main` branch. 

However, you can also build it locally:

```bash
docker build -t kisbogdan/gallery:latest .
docker push kisbogdan/gallery:latest
```

> Warning: Currently, the env variables are burned into the `deploy.yaml`, so the deployment is **not** production-ready.

### Kubernetes

Use the `deploy.yaml` to deploy to Kubernetes.

```bash
kubectl apply -f deploy.yaml
```

### OpenShift (BME Cloud)

Modify the `deploy.yaml`.

1. remove the `PersistetVolume` sections (Postgres and SeaweedFS)
1. modify the `PersistentVolumeClaim` `storageClassName` from `manual` to `rook-cephfs` (Postgres and SeaweedFS)

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

Each endpoint can return the documented data type in success, or an error message, like:

```json
{
  "error": "image not found"
}
```

The images are stored and retrieved separately from the JSON data, to be able to use JSON for all the requests. See endpoints later.

#### POST `/api/user/register`

Register to the service.

Request body:

```json
{
  "username": "admin",
  "password": "Almafa12",
  "displayname": "Admin"
}
```

Reply:

```json
{
  "username": "admin",
  "id": 1,
  "displayname": "Admin",
  "registered": "2026-03-15T10:04:49.559563836Z",
  "admin": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzM1NzA4ODksInVzZXJpZCI6MSwiYWRtaW4iOmZhbHNlfQ.LCVVBN9yW1QbhHvwm-75kiEy9SHNZ5DcTcud73qDVwo"
}
```

#### POST `/api/user/login`

Logs in the user.

Request body:

```json
{
  "username": "admin",
  "password": "Almafa12"
}
```

Reply:

```json
{
  "username": "admin",
  "id": 1,
  "displayname": "Admin",
  "registered": "2026-03-02T07:05:42.002978Z",
  "admin": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzM1NzMwMjUsInVzZXJpZCI6MSwiYWRtaW4iOmZhbHNlfQ.lITlXB1azc2PbAD-K7Z5bjbyFTMxzjyGkDLqhTlJNbw"
}
```

#### GET `/api/user/me`

Returns the currently logged-in user.

Reply:

```json
{
  "username": "admin",
  "id": 1,
  "displayname": "Admin",
  "registered": "2026-02-26T17:30:24.696482+01:00",
  "admin": true
}
```

#### GET `/api/user/all`

Returns all the users. Only allowed for admins.

Reply:

```json
[
  {
    "username": "admin",
    "id": 1,
    "displayname": "Admin",
    "registered": "2026-03-02T07:05:42.002978Z",
    "admin": false
  }
]
```

#### POST `/api/image`

Create an image. Uploading the actual image is done in a different request to not mess up the JSON.

Request body:

```json
{
  "name": "Szép kép 4"
}
```

Reply:

```json
{
  "id": 13,
  "name": "Szép kép 4",
  "uploaded": "2026-03-15T11:29:52.304718923+01:00",
  "uploader": {
    "username": "admin",
    "id": 1,
    "displayname": "Admin",
    "registered": "2026-02-26T17:30:24.696482+01:00",
    "admin": true
  },
  "image": null
}
```

#### POST `/api/image/:id/upload`

Upload the image data for given ID. Only the creator of an image can upload or an admin, and only if there is no image data uploaded.

Request body: Raw image data

Reply:

```json
{
  "id": 13,
  "name": "Szép kép 4",
  "uploaded": "2026-03-15T11:29:52.304718+01:00",
  "uploader": {
    "username": "admin",
    "id": 1,
    "displayname": "Admin",
    "registered": "2026-02-26T17:30:24.696482+01:00",
    "admin": true
  },
  "image": "00a2a3f4-f2db-4385-a651-166c0b91d6d0"
}
```

#### GET `/api/image`

Returns all the images in the system.

Reply:

```json
[
  {
    "id": 2,
    "name": "Szép kép 2.0",
    "uploaded": "2026-02-26T21:50:37.425539+01:00",
    "uploader": {
      "username": "admin",
      "id": 1,
      "displayname": "Admin",
      "registered": "2026-02-26T17:30:24.696482+01:00",
      "admin": true
    },
    "image": "b692592a-dd91-4a11-ac88-c81d06cead65"
  },
  {
    "id": 3,
    "name": "Szép kép 3.0",
    "uploaded": "2026-03-15T11:12:47.086573+01:00",
    "uploader": {
      "username": "admin",
      "id": 1,
      "displayname": "Admin",
      "registered": "2026-02-26T17:30:24.696482+01:00",
      "admin": true
    },
    "image": "a4542c99-3697-435f-a6c3-7228c1f7492f"
  }
]
```

#### DELETE `/api/image/:id`

Delete image with given ID. Only the creator of the image or an admin can delete.

Reply:

```json
{
  "status": "deleted"
}
```

#### GET `/api/storage/:id`

Returns the image with the given UUID from the image list.
