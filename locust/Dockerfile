# syntax=docker/dockerfile:1

FROM locustio/locust

# Copy the locustfile and the images directory into the container
COPY locustfile.py /mnt/locust/locustfile.py
COPY --chmod=755 images/ /mnt/locust/images/ 

WORKDIR /mnt/locust