import os
import random
import string
from locust import HttpUser, task, between

IMAGE_DIR = os.path.join(os.path.dirname(__file__), "images")

def generate_random_string(length=10):
    """Generate a random alphanumeric string."""
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

class GalleryUser(HttpUser):
    wait_time = between(1, 5)

    def on_start(self):
        """Create a new user and login on startup."""
        self.username = f"user_{generate_random_string()}"
        self.password = "Secret123!"
        self.displayname = f"Test User {self.username}"
        self.token = None
        self.headers = {}
        self.my_images = []

        # Register the user
        register_resp = self.client.post("/api/user/register", json={
            "username": self.username,
            "password": self.password,
            "displayname": self.displayname
        })

        if register_resp.status_code == 200:
            data = register_resp.json()
            if "token" in data:
                self.token = data["token"]
                self.headers["Authorization"] = f"Bearer {self.token}"

    @task(3)
    def view_all_images(self):
        """View the list of all images and download their data."""
        resp = self.client.get("/api/image")
        if resp.status_code == 200:
            images = resp.json()
            random.shuffle(images)
            for img in images[:20]:
                storage_uuid = img.get("image")
                if storage_uuid:
                    self.client.get(f"/api/storage/{storage_uuid}", name="/api/storage/[uuid]")

    @task(1)
    def view_profile(self):
        """View the current user's profile."""
        if self.token:
            self.client.get("/api/user/me", headers=self.headers)

    @task(1)
    def periodic_login(self):
        """Periodically login to simulate authentication traffic."""
        self.client.post("/api/user/login", json={
            "username": self.username,
            "password": self.password
        })

    @task(2)
    def upload_new_image(self):
        """Create a new image record and upload data to it."""
        if not self.token:
            return

        # 1. Create image record
        image_name = f"Locust Image {generate_random_string(5)}"
        create_resp = self.client.post("/api/image", json={
            "name": image_name
        }, headers=self.headers)

        if create_resp.status_code == 200:
            image_data = create_resp.json()
            image_id = image_data.get("id")

            if image_id:
                # 2. Upload actual image data from 'image' directory
                upload_headers = self.headers.copy()
                upload_headers["Content-Type"] = "image/jpeg"
                
                try:
                    images_available = os.listdir(IMAGE_DIR)
                    if images_available:
                        selected_image = random.choice(images_available)
                        image_path = os.path.join(IMAGE_DIR, selected_image)
                        with open(image_path, "rb") as f:
                            payload = f.read()
                    else:
                        payload = b"dummy image content"
                except FileNotFoundError:
                    payload = b"dummy image content"

                upload_resp = self.client.post(
                    f"/api/image/{image_id}/upload", 
                    data=payload, 
                    headers=upload_headers,
                    name="/api/image/[id]/upload"
                )

                if upload_resp.status_code == 200:
                    self.my_images.append(image_id)
                    
                    # 3. View the uploaded image through storage API
                    storage_uuid = upload_resp.json().get("image")
                    if storage_uuid:
                        self.client.get(f"/api/storage/{storage_uuid}", name="/api/storage/[uuid]")

    @task(1)
    def delete_random_image(self):
        """Delete one of the user's uploaded images."""
        if not self.token or not self.my_images:
            return

        image_id = random.choice(self.my_images)
        delete_resp = self.client.delete(f"/api/image/{image_id}", headers=self.headers, name="/api/image/[id]")
        
        if delete_resp.status_code == 200:
            self.my_images.remove(image_id)

    def on_stop(self):
        """Clean up all the created images when the simulated user stops."""
        if not self.token or not self.my_images:
            return

        for image_id in list(self.my_images):
            delete_resp = self.client.delete(f"/api/image/{image_id}", headers=self.headers, name="/api/image/[id]")
            if delete_resp.status_code == 200:
                self.my_images.remove(image_id)
