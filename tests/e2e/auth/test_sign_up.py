from httpx import Client


def test_sign_up(client: Client):
    """Tests the sign-up endpoint by creating a new user."""
    r = client.post("/auth/sign-up", json={
                "email": "john-doe@mail.com",
                "password": "Secret1234",
                "username": "john-doe12"
        })
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["access_token"] is not None
    assert body["account"]["id"] is not None
    assert body["account"]["email"] == "john-doe@mail.com"

def test_sign_up_existing_email_return_409(client: Client):
    """Tests that signing up with an existing email returns a 409 error."""
    r = client.post("/auth/sign-up", json={
                "email": "john-doe@mail.com",
                "password": "Secret1234",
                "username": "john-doe12"
        })
    assert r.status_code == 409, r.text
    body = r.json()
    assert body["error"] == "Email already registered"

def test_sign_up_incorrect_email_return_400(client: Client):
    """Tests that signing up with incorrect email returns a 400 error."""
    r = client.post("/auth/sign-up", json={
                "email": "invalid-email",
                "password": "Secret1234",
                "username": "john-doe12"
        })

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "Must be a valid email address"

def test_sign_up_short_password_return_400(client: Client):
    """Tests that signing up with a short password returns a 400 error."""
    r = client.post("/auth/sign-up", json={
                "email": "john-doe@mail.com",
                "password": "Short1",
                "username": "john-doe12"
        })

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["password"] == "Must be at least 8 characters"

def test_sign_up_missing_fields_return_400(client: Client):
    """Tests that signing up with missing fields returns a 400 error."""
    r = client.post("/auth/sign-up", json={})
    assert r.status_code == 400, r.text
    body = r.json()
    print(body)

    assert body["error"] == "Validation failed"
    assert body["fields"]["email"] == "This field is required"
    assert body["fields"]["password"] == "This field is required"
    assert body["fields"]["username"] == "This field is required"