import pytest
from httpx import Client
from e2e.helpers import create_test_png

USER = {"email": "post-guild-tester@mail.com", "username": "post-guild-tester", "password": "Secret1234"}

@pytest.fixture(scope="module")
def user_payload():
    return USER

def test_post_guild(auth_client: Client, sign_in_user: dict):
    r = auth_client.post("/guilds", data={"name": "Test Guild", "description": "A guild created during testing"})
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["name"] == "Test Guild"
    assert body["owner_id"] == str(sign_in_user["id"])

def test_post_guild_with_icon(auth_client: Client, sign_in_user: dict):
    files = {"icon": ("icon.png", create_test_png(), "image/png")}
    r = auth_client.post("/guilds", data={"name": "Guild with Icon"}, files=files)
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["icon"] is not None
    assert body["owner_id"] == str(sign_in_user["id"])

def test_post_guild_unauthorized(client: Client):
    r = client.post("/guilds", data={"name": "Test Guild"})
    assert r.status_code == 401, r.text

def test_post_guild_missing_name(auth_client: Client):
    r = auth_client.post("/guilds", data={})
    assert r.status_code == 400, r.text
    assert r.json()["fields"]["name"] == "This field is required"

def test_post_guild_name_too_long(auth_client: Client):
    r = auth_client.post("/guilds", data={"name": "A" * 101})
    assert r.status_code == 400, r.text
    assert r.json()["fields"]["name"] == "Must be at most 100 characters"

def test_post_guild_description_too_long(auth_client: Client):
    r = auth_client.post("/guilds", data={"name": "Test Guild", "description": "A" * 301})
    assert r.status_code == 400, r.text
    assert r.json()["fields"]["description"] == "Must be at most 300 characters"