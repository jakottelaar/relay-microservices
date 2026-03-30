import pytest
from httpx import Client

USER = {"email": "post-guild-channel-tester@mail.com", "username": "post-guild-channel-tester", "password": "Secret1234"}

@pytest.fixture(scope="module")
def user_payload():
    return USER

@pytest.fixture(scope="module")
def guild(auth_client: Client):
    r = auth_client.post("/guilds", data={"name": "Test Guild", "description": "A guild created during testing"})
    assert r.status_code == 201, r.text
    return r.json()

def test_post_guild_channel(auth_client: Client, guild: dict):
    r = auth_client.post(f"/guilds/{guild['id']}/channels", json={"name": "general", "type": 1, "topic": "General discussion"})
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["name"] == "general"
    assert body["guild_id"] == guild["id"]
    assert body["type"] == 1
    assert body["topic"] == "General discussion"

def test_post_guild_channel_without_type(auth_client: Client, guild: dict):
    r = auth_client.post(f"/guilds/{guild['id']}/channels", json={"name": "no-type-channel"})
    assert r.status_code == 201, r.text
    assert r.json()["type"] == 1

def test_post_guild_channel_unauthorized(client: Client, guild: dict):
    r = client.post(f"/guilds/{guild['id']}/channels", json={"name": "general", "type": 1})
    assert r.status_code == 401, r.text

def test_post_guild_channel_missing_name(auth_client: Client, guild: dict):
    r = auth_client.post(f"/guilds/{guild['id']}/channels", json={"type": 1})
    assert r.status_code == 400, r.text
    assert r.json()["fields"]["name"] == "This field is required"

def test_post_guild_channel_invalid_type(auth_client: Client, guild: dict):
    r = auth_client.post(f"/guilds/{guild['id']}/channels", json={"name": "general", "type": 99})
    assert r.status_code == 400, r.text

def test_post_guild_channel_nonexistent_guild(auth_client: Client):
    r = auth_client.post("/guilds/9999/channels", json={"name": "general", "type": 1})
    assert r.status_code == 404, r.text
    assert r.json()["error"] == "Guild not found"