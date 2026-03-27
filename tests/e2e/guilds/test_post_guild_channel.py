from httpx import Client
import pytest

USER = {
    "email": "post-guild-tester@mail.com",
    "username": "post-guild-tester",
    "password": "Secret1234",
}

GUILD = {
    "name": "Test Guild",
    "description": "A guild created during testing"
}

@pytest.fixture(scope="module", autouse=True)
def setup_test_account(client: Client):
    client.post("/auth/sign-up", json=USER)
    yield

@pytest.fixture(scope="module")
def setup_test_guild(auth_client: Client):
    
    r = auth_client.post("/guilds", data=GUILD)
    assert r.status_code == 201, r.text
    return r.json()

@pytest.fixture(scope="module")
def user_payload():
    return USER

def test_post_guild_channel(auth_client: Client, setup_test_guild):
    guild_id = setup_test_guild["id"]
    r = auth_client.post(f"/guilds/{guild_id}/channels", json={"name": "general", "type": 1, "topic": "General discussion"})

    assert r.status_code == 201, r.text
    body = r.json()
    assert body["id"] is not None
    assert body["name"] == "general"
    assert body["guild_id"] == guild_id
    assert body["type"] == 1
    assert body["topic"] == "General discussion"

def test_post_guild_channel_without_type(auth_client: Client, setup_test_guild):
    guild_id = setup_test_guild["id"]
    r = auth_client.post(f"/guilds/{guild_id}/channels", json={"name": "general"})

    assert r.status_code == 201, r.text
    body = r.json()
    assert body["id"] is not None
    assert body["name"] == "general"
    assert body["guild_id"] == guild_id
    assert body["type"] == 1  # Default to text channel

def test_post_guild_channel_unauthorized(client: Client, setup_test_guild):
    guild_id = setup_test_guild["id"]
    r = client.post(f"/guilds/{guild_id}/channels", json={"name": "general", "type": 1})

    assert r.status_code == 401, r.text

def test_post_guild_channel_missing_name(auth_client: Client, setup_test_guild):
    guild_id = setup_test_guild["id"]
    r = auth_client.post(f"/guilds/{guild_id}/channels", json={"type": 1})

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["name"] == "This field is required"

def test_post_guild_channel_invalid_type(auth_client: Client, setup_test_guild):
    guild_id = setup_test_guild["id"]
    r = auth_client.post(f"/guilds/{guild_id}/channels", json={"name": "general", "type": 99})

    assert r.status_code == 400, r.text
    body = r.json()
    assert body["error"] == "Validation failed"
    assert body["fields"]["type"] == "Invalid value"

def test_post_guild_channel_nonexistent_guild(auth_client: Client):
    r = auth_client.post(f"/guilds/9999/channels", json={"name": "general", "type": 1})

    assert r.status_code == 404, r.text
    body = r.json()
    assert body["error"] == "Guild not found"