from httpx import Client
import pytest

USER = {
    "email": "get-user-tester@mail.com",
    "username": "get-user-tester",
    "password": "Secret1234",
}

@pytest.fixture(scope="module", autouse=True)
def setup_test_account(client: Client):
    client.post("/auth/sign-up", json=USER)
    yield

@pytest.fixture(scope="module")
def user_payload():
    return USER

def test_get_user_profile(auth_client: Client, sign_in_user):
    user_id = sign_in_user["account"]["id"]

    r = auth_client.get(f"/users/{user_id}/profile")

    assert r.status_code == 200, r.text
    body = r.json()

    assert body["id"] == user_id
    assert body["username"] == USER["username"]

def test_get_user_profile_unauthorized(client: Client, sign_in_user):
    user_id = sign_in_user["account"]["id"]

    r = client.get(f"/users/{user_id}/profile")

    assert r.status_code == 401, r.text

def test_get_user_profile_non_existent_user(auth_client: Client):
    r = auth_client.get("/users/999999/profile")

    assert r.status_code == 404, r.text
    body = r.json()
    assert body["error"] == "user not found"