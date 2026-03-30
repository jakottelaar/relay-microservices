import pytest
from httpx import Client

USER = {"email": "get-user-profile-tester@mail.com", "username": "get-user-profile-tester", "password": "Secret1234"}

@pytest.fixture(scope="module")
def user_payload():
    return USER

def test_get_user_profile(auth_client: Client, sign_in_user: dict, user_payload: dict):
    r = auth_client.get(f"/users/{sign_in_user['id']}/profile")
    assert r.status_code == 200, r.text
    body = r.json()
    assert body["id"] == sign_in_user["id"]
    assert body["username"] == user_payload["username"]

def test_get_user_profile_unauthorized(client: Client, sign_in_user: dict):
    r = client.get(f"/users/{sign_in_user['id']}/profile")
    assert r.status_code == 401, r.text

def test_get_user_profile_non_existent_user(auth_client: Client):
    r = auth_client.get("/users/999999/profile")
    assert r.status_code == 404, r.text
    assert r.json()["error"] == "user not found"