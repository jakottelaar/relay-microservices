from httpx import Client
import pytest
from PIL import Image
import io

USER = {
    "email": "post-guild-tester@mail.com",
    "username": "post-guild-tester",
    "password": "Secret1234",
}

@pytest.fixture(scope="module", autouse=True)
def setup_test_account(client: Client):
    client.post("/auth/sign-up", json=USER)
    yield

@pytest.fixture(scope="module")
def user_payload():
    return USER

def create_test_png() -> bytes:
    img = Image.new("RGB", (1, 1), color=(255, 0, 0))
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()

def test_post_guild(auth_client: Client, sign_in_user):
    r = auth_client.post(
        "/guilds",
        data={
            "name": "Test Guild",
            "description": "A guild created during testing"
        }
    )

    assert r.status_code == 201, r.text
    body = r.json()

    assert body["name"] == "Test Guild"
    assert body["owner_id"] == str(sign_in_user["account"]["id"])

def test_post_guild_with_icon(auth_client: Client, sign_in_user):

    png_data = create_test_png()
    files = {"icon": ("icon.png", png_data, "image/png")}
    r = auth_client.post("/guilds", data={"name": "Guild with Icon"}, files=files)

    assert r.status_code == 201, r.text
    body = r.json()
    assert body["name"] == "Guild with Icon"
    assert body["icon"] is not None
    assert body["owner_id"] == str(sign_in_user["account"]["id"])

def test_post_guild_unauthorized(client: Client):
    r = client.post("/guilds", data={"name": "Test Guild"})

    assert r.status_code == 401, r.text

def test_post_guild_missing_name(auth_client: Client):
    r = auth_client.post("/guilds", data={})

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["name"] == "This field is required"

def test_post_guild_name_too_long(auth_client: Client):
    r = auth_client.post("/guilds", data={"name": "A" * 101})

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["name"] == "Must be at most 100 characters"

def test_post_guild_description_too_long(auth_client: Client):
    r = auth_client.post("/guilds", data={"name": "Test Guild", "description": "A" * 501})

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["description"] == "Must be at most 500 characters"

