import pytest
from httpx import Client
from e2e.helpers import sign_up, sign_in, create_guild, create_channel

OWNER = {"email": "guild-owner@relay.dev", "username": "guild-owner", "password": "Secret1234"}

@pytest.fixture(scope="module")
def owner(client: Client):
    sign_up(client, OWNER)
    auth, _ = sign_in(client, OWNER["email"], OWNER["password"])
    yield auth
    auth.close()

@pytest.fixture(scope="module")
def guild(owner: Client):
    return create_guild(owner, "Test Guild")

@pytest.fixture(scope="module")
def channels(owner: Client, guild: dict):
    channel1 = create_channel(owner, guild["id"], "general")
    channel2 = create_channel(owner, guild["id"], "random")
    return [channel1, channel2]

def test_get_guild_channels(owner: Client, guild: dict, channels: list):
    r = owner.get(f"/guilds/{guild['id']}/channels")
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) == len(channels)
    channel_ids = {channel["id"] for channel in channels}
    for channel in body:
        assert channel["id"] in channel_ids

def test_get_guild_channels_unauthorized(client: Client, guild: dict):
    r = client.get(f"/guilds/{guild['id']}/channels")
    assert r.status_code == 401, r.text

def test_get_guild_channels_guild_not_found(owner: Client):
    r = owner.get(f"/guilds/123345345/channels")
    assert r.status_code == 404, r.text
    body = r.json()
    assert body["error"] == "Guild not found"

def test_get_guild_channels_no_channels(owner: Client):
    empty_guild = create_guild(owner, "Empty Guild")
    r = owner.get(f"/guilds/{empty_guild['id']}/channels")
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) == 0