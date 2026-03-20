import pytest
import httpx
import os
from dotenv import load_dotenv

load_dotenv()

BASE_URL = os.getenv("API_BASE_URL", "http://localhost:80")

@pytest.fixture(scope="session")
def client() -> httpx.Client:
    with httpx.Client(base_url=BASE_URL) as client:
        yield client

@pytest.fixture
def auth_client(signed_in_user, client) -> httpx.Client:
    token = signed_in_user["access_token"]
    return httpx.Client(
        base_url=BASE_URL,
        headers={"Authorization": f"Bearer {token}"},
        timeout=10
    )

@pytest.fixture
def sign_in_user(client, user_payload):
    """Signs in a user and returns the full response body."""
    r = client.post("/api/auth/sign-in", json=user_payload)
    assert r.status_code == 200, r.text
    return r.json()
