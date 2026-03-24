from httpx import Client
import pytest

USER = {
    "email": "sign-in-tester@mail.com",
    "username": "sign-in-tester",
    "password": "Secret1234",
}

@pytest.fixture(scope="module", autouse=True)
def setup_test_account(client: Client):
    client.post("/auth/sign-up", json=USER)
    yield

@pytest.fixture(scope="module")
def user_payload():
    return USER

def test_sign_in(sign_in_user):
    body = sign_in_user

    assert body["access_token"] is not None
    assert body["account"]["id"] is not None

def test_sign_in_incorrect_password_return_401(client: Client):
    """Tests that signing in with an incorrect password returns a 401 error."""
    r = client.post("/auth/sign-in", json={
        "email": "sign-in-tester@mail.com",
        "password": "Secret9999",
    })

    assert r.status_code == 401, r.text
    body = r.json()
    assert body["error"] == "invalid email or password"

def test_sign_in_non_existent_email_return_401(client: Client):
    """Tests that signing in with a non-existent email returns a 401 error."""
    r = client.post("/auth/sign-in", json={
                "email": "non-existent@mail.com",
                "password": "Secret1234",
    })

    assert r.status_code == 401, r.text
    body = r.json()
    assert body["error"] == "invalid email or password"

def test_sign_in_missing_email_return_400(client: Client):
    """Tests that signing in with missing email returns a 400 error."""
    r = client.post("/auth/sign-in", json={
            "password": "Secret1234",
    })

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "This field is required"

def test_sign_in_missing_password_return_400(client: Client):
    """Tests that signing in with missing password returns a 400 error."""
    r = client.post("/auth/sign-in", json={
            "email": "sign-in-tester@mail.com",
    })

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["password"] == "This field is required"

def test_sign_in_invalid_email_format_return_400(client: Client):
    """Tests that signing in with an invalid email format returns a 400 error."""
    r = client.post("/auth/sign-in", json={
            "email": "invalid-email-format",
            "password": "Secret1234",
    })

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "Must be a valid email address"