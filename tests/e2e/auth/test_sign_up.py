from httpx import Client

def test_sign_up(client: Client):
    r = client.post("/auth/sign-up", json={"email": "john-doe@mail.com", "password": "Secret1234", "username": "john-doe12"})
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["access_token"] is not None
    assert body["account"]["id"] is not None
    assert body["account"]["email"] == "john-doe@mail.com"

def test_sign_up_existing_email_return_409(client: Client):
    r = client.post("/auth/sign-up", json={"email": "john-doe@mail.com", "password": "Secret1234", "username": "john-doe12"})
    assert r.status_code == 409, r.text
    assert r.json()["error"] == "Email already registered"

def test_sign_up_incorrect_email_return_400(client: Client):
    r = client.post("/auth/sign-up", json={"email": "invalid-email", "password": "Secret1234", "username": "john-doe12"})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "Must be a valid email address"

def test_sign_up_short_password_return_400(client: Client):
    r = client.post("/auth/sign-up", json={"email": "john-doe@mail.com", "password": "Short1", "username": "john-doe12"})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["password"] == "Must be at least 8 characters"

def test_sign_up_missing_fields_return_400(client: Client):
    r = client.post("/auth/sign-up", json={})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "This field is required"
    assert body["fields"]["password"] == "This field is required"
    assert body["fields"]["username"] == "This field is required"