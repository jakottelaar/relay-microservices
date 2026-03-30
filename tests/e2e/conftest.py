import pytest
import httpx
import os
from dotenv import load_dotenv
from e2e.helpers import sign_up, sign_in

load_dotenv()

BASE_URL = os.getenv("API_BASE_URL", "http://localhost:80")


@pytest.fixture(scope="session")
def client() -> httpx.Client:
    with httpx.Client(base_url=BASE_URL, timeout=10) as c:
        yield c


@pytest.fixture(scope="module")
def auth_client(client: httpx.Client, user_payload: dict):
    sign_up(client, user_payload)
    auth, _ = sign_in(client, user_payload["email"], user_payload["password"])
    yield auth
    auth.close()


@pytest.fixture(scope="module")
def sign_in_user(auth_client: httpx.Client, client: httpx.Client, user_payload: dict):
    _, data = sign_in(client, user_payload["email"], user_payload["password"])
    return data