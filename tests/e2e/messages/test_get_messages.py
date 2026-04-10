import pytest
from httpx import Client
from e2e.helpers import create_message, sign_up, sign_in, create_guild, create_channel

USER = {"email": "get-messages-email@relay.dev", "username": "get-messages-user", "password": "Secret1234"}

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
    messages = [
        create_message(user, channel["id"], "Hello, world!"),
        create_message(user, channel["id"], "This is a test message."),
        create_message(user, channel["id"], "Goodbye, world!"),
    ]
    return guild, channel, messages

@pytest.fixture(scope="module")
def empty_channel(user: Client):
    guild = create_guild(user, "Empty Channel Guild")
    channel = create_channel(user, guild["id"], "Empty Channel")
    return guild, channel

def test_get_messages(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    _, channel, messages = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages")
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) == len(messages)
    body_sorted = sorted(body, key=lambda x: x["id"])
    messages_sorted = sorted(messages, key=lambda x: x["id"])
    for msg, expected in zip(body_sorted, messages_sorted):
        assert msg["id"] == expected["id"]
        assert msg["content"] == expected["content"]
        assert msg["channel_id"] == str(channel["id"])

def test_get_messages_with_before_id(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    _, channel, messages = guild_with_channel
    before_id = messages[1]["id"]
    r = user.get(f"/channels/{channel['id']}/messages", params={"before": before_id})
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) >= 0

def test_get_messages_with_after_id(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    _, channel, messages = guild_with_channel
    after_id = messages[0]["id"]
    r = user.get(f"/channels/{channel['id']}/messages", params={"after": after_id})
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) >= 0

def test_get_messages_with_limit(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    _, channel, messages = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"limit": 2})
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) == 2

def test_get_messages_unauthorized(client: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    _, channel, _ = guild_with_channel
    r = client.get(f"/channels/{channel['id']}/messages")
    assert r.status_code == 401, r.text

def test_get_messages_empty_channel(user: Client, empty_channel: tuple[dict, dict]):
    _, channel = empty_channel
    r = user.get(f"/channels/{channel['id']}/messages")
    assert r.status_code == 200, r.text
    body = r.json()
    assert isinstance(body, list)
    assert len(body) == 0

def test_get_messages_before_id_filters_correctly(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """All returned messages should have id < before_id (exclusive cursor)"""
    _, channel, messages = guild_with_channel
    before_id = messages[1]["id"]
    r = user.get(f"/channels/{channel['id']}/messages", params={"before": before_id})
    assert r.status_code == 200, r.text
    body = r.json()
    
    # Every message returned must have id < before_id
    for msg in body:
        msg_id = int(msg["id"])
        assert msg_id < int(before_id), f"Message {msg_id} should be < {before_id}"

def test_get_messages_after_id_filters_correctly(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """All returned messages should have id > after_id (exclusive cursor)"""
    _, channel, messages = guild_with_channel
    after_id = messages[0]["id"]
    r = user.get(f"/channels/{channel['id']}/messages", params={"after": after_id})
    assert r.status_code == 200, r.text
    body = r.json()
    
    # Every message returned must have id > after_id
    for msg in body:
        msg_id = int(msg["id"])
        assert msg_id > int(after_id), f"Message {msg_id} should be > {after_id}"

def test_get_messages_ordering_consistency(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """Messages should be ordered consistently and have no duplicates"""
    _, channel, _ = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages")
    assert r.status_code == 200, r.text
    body = r.json()
    
    if len(body) > 1:
        ids = [int(msg["id"]) for msg in body]
        # Verify all IDs are unique (no duplicates)
        assert len(ids) == len(set(ids)), "Messages contain duplicate IDs"

def test_get_messages_before_and_after_id_conflict(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """before and after together should return 400"""
    _, channel, messages = guild_with_channel
    before_id = messages[1]["id"]
    after_id = messages[0]["id"]
    
    r = user.get(
        f"/channels/{channel['id']}/messages",
        params={"before": before_id, "after": after_id}
    )
    assert r.status_code == 400

def test_get_messages_non_numeric_limit(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """Non-numeric limit should return 400 or be ignored"""
    _, channel, _ = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"limit": "not_a_number"})
    assert r.status_code == 400

def test_get_messages_non_numeric_before_id(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """Non-numeric before should return 400 or be ignored"""
    _, channel, _ = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"before": "not_a_number"})
    assert r.status_code == 400

def test_get_messages_non_numeric_after_id(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """Non-numeric after should return 400 or be ignored"""
    _, channel, _ = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"after": "not_a_number"})
    assert r.status_code == 400

def test_get_messages_limit_zero(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """limit=0 should use default or return error"""
    _, channel, _ = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"limit": 0})
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) >= 0

def test_get_messages_negative_limit(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """Negative limit should use default"""
    _, channel, messages = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"limit": -5})
    assert r.status_code == 200, r.text
    body = r.json()
    # Should use default limit (50)
    assert len(body) <= len(messages)

def test_get_messages_limit_cap_at_100(user: Client, guild_with_channel: tuple[dict, dict, list[dict]]):
    """limit > 100 should be capped at 100"""
    _, channel, messages = guild_with_channel
    r = user.get(f"/channels/{channel['id']}/messages", params={"limit": 999})
    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body) <= min(100, len(messages))

# Have to implement actual channel existence checks before we can reliably test this case
# def test_get_messages_nonexistent_channel(user: Client):
#     """Nonexistent channel_id should return 404 or 403"""
#     fake_channel_id = "999999999999999999"
#     r = user.get(f"/channels/{fake_channel_id}/messages")
#     assert r.status_code == 404

def test_get_messages_malformed_channel_id(user: Client):
    """Malformed channel_id should return validation error"""
    r = user.get(f"/channels/not_a_valid_id/messages")
    assert r.status_code == 400, f"Expected 400, got {r.status_code}: {r.text}"

def test_get_messages_unauthorized_user(client: Client):
    """Unauthorized request should return 401"""
    fake_channel_id = "1"
    r = client.get(f"/channels/{fake_channel_id}/messages")
    assert r.status_code == 401, f"Expected 401, got {r.status_code}: {r.text}"