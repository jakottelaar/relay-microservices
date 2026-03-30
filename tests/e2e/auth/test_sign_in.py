import pytest
from httpx import Client

USER = {"email": "sign-in-tester@mail.com", "username": "sign-in-tester", "password": "Secret1234"}

@pytest.fixture(scope="module")
def user_payload():
    return USER

def test_sign_in(sign_in_user: dict):
    assert sign_in_user["id"] is not None

def test_sign_in_incorrect_password_return_401(client: Client, user_payload: dict):
    r = client.post("/auth/sign-in", json={"email": user_payload["email"], "password": "Secret9999"})
    assert r.status_code == 401, r.text
    assert r.json()["error"] == "invalid email or password"

def test_sign_in_non_existent_email_return_401(client: Client):
    r = client.post("/auth/sign-in", json={"email": "non-existent@mail.com", "password": "Secret1234"})
    assert r.status_code == 401, r.text
    assert r.json()["error"] == "invalid email or password"

def test_sign_in_missing_email_return_400(client: Client):
    r = client.post("/auth/sign-in", json={"password": "Secret1234"})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "This field is required"

def test_sign_in_missing_password_return_400(client: Client, user_payload: dict):
    r = client.post("/auth/sign-in", json={"email": user_payload["email"]})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["password"] == "This field is required"

def test_sign_in_invalid_email_format_return_400(client: Client):
    r = client.post("/auth/sign-in", json={"email": "invalid-email-format", "password": "Secret1234"})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "Must be a valid email address"