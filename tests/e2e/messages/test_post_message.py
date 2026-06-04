import pytest
from httpx import Client
from e2e.helpers import sign_up, sign_in, create_guild, create_channel

USER = {"email": "post-messages-email@relay.dev", "username": "post-messages-user", "password": "Secret1234"}

@pytest.fixture(scope="module")
def user(client: Client):
    sign_up(client, USER)
    auth, _ = sign_in(client, USER["email"], USER["password"])
    yield auth
    auth.close()

@pytest.fixture(scope="module")
def guild_with_channel(user: Client):
    guild = create_guild(user, "Messages Guild")
    channel = create_channel(user, guild["id"], "Messages Channel")
    return guild, channel

def test_post_message(user: Client, guild_with_channel: tuple[dict, dict]):
    _, channel = guild_with_channel
    r = user.post(f"/channels/{channel['id']}/messages", json={"content": "Hello, world!"})
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["content"] == "Hello, world!"
    assert body["channel_id"] == str(channel["id"])

def test_post_message_empty_content(user: Client, guild_with_channel: tuple[dict, dict]):
    _, channel = guild_with_channel
    r = user.post(f"/channels/{channel['id']}/messages", json={"content": ""})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["fields"]["content"] == "Must be at least 1 character"

def test_post_message_missing_content(user: Client, guild_with_channel: tuple[dict, dict]):
    _, channel = guild_with_channel
    r = user.post(f"/channels/{channel['id']}/messages", json={})
    assert r.status_code == 400, r.text
    body = r.json()
    assert body["fields"]["content"] == "This field is required"

def test_post_message_unauthorized(client: Client, guild_with_channel: tuple[dict, dict]):
    _, channel = guild_with_channel
    r = client.post(f"/channels/{channel['id']}/messages", json={"content": "Hello, world!"})
    assert r.status_code == 401, r.text
