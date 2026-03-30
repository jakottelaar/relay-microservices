import pytest
from httpx import Client
from e2e.helpers import create_test_png

USER = {"email": "get-guild-tester@mail.com", "username": "get-guild-tester", "password": "Secret1234"}

@pytest.fixture(scope="module")
def user_payload():
    return USER

@pytest.fixture(scope="module")
def guild(auth_client: Client):
    files = {"icon": ("icon.png", create_test_png(), "image/png")}
    r = auth_client.post("/guilds", data={"name": "Test Guild", "description": "A guild for testing"}, files=files)
    assert r.status_code == 201, r.text
    return r.json()

def test_get_guild(auth_client: Client, guild: dict):
    r = auth_client.get(f"/guilds/{guild['id']}")
    assert r.status_code == 200, r.text
    body = r.json()
    assert body["id"] == guild["id"]
    assert body["name"] == "Test Guild"
    assert body["description"] == "A guild for testing"
    assert body["icon"] is not None

def test_get_guild_not_found(auth_client: Client):
    r = auth_client.get("/guilds/123453454564564564")
    assert r.status_code == 404, r.text
    assert r.json()["error"] == "Guild not found"

def test_get_guild_unauthenticated(client: Client):
    r = client.get("/guilds/123453454564564564")
    assert r.status_code == 401, r.text