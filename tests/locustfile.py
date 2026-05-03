from locust import HttpUser, task, between


class SentinelUser(HttpUser):
    wait_time = between(0.01, 0.2)

    @task(8)
    def light(self):
        self.client.get("/")

    @task(2)
    def heavy(self):
        self.client.get("/heavy")
