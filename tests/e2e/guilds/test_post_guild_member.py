import pytest
from httpx import Client
from e2e.helpers import sign_up, sign_in, create_guild

OWNER = {"email": "guild-owner@relay.dev", "username": "guild-owner", "password": "Secret1234"}
MEMBER = {"email": "guild-member@relay.dev", "username": "guild-member", "password": "Secret1234"}

@pytest.fixture(scope="module")
def owner(client: Client):
    sign_up(client, OWNER)
    auth, _ = sign_in(client, OWNER["email"], OWNER["password"])
    yield auth
    auth.close()

@pytest.fixture(scope="module")
def owner_account(client: Client):
    _, account = sign_in(client, OWNER["email"], OWNER["password"])
    return account

@pytest.fixture(scope="module")
def member(client: Client):
    sign_up(client, MEMBER)
    auth, _ = sign_in(client, MEMBER["email"], MEMBER["password"])
    yield auth
    auth.close()

@pytest.fixture(scope="module")
def member_account(client: Client):
    sign_up(client, MEMBER)
    _, account = sign_in(client, MEMBER["email"], MEMBER["password"])
    return account

@pytest.fixture(scope="module")
def guild(owner: Client):
    return create_guild(owner, "Test Guild")


def test_add_member(owner: Client, guild: dict, member_account: dict):
    r = owner.post(f"/guilds/{guild['id']}/members/{member_account['id']}", json={})
    assert r.status_code == 201, r.text
    body = r.json()
    assert body["user_id"] == str(member_account["id"])

def test_add_member_already_member(owner: Client, guild: dict, member_account: dict):
    r = owner.post(f"/guilds/{guild['id']}/members/{member_account['id']}", json={})
    assert r.status_code == 409, r.text
    assert r.json()["error"] == "User is already a member of this guild"

def test_add_owner_as_member_returns_error(owner: Client, guild: dict, owner_account: dict):
    r = owner.post(f"/guilds/{guild['id']}/members/{owner_account['id']}", json={})
    assert r.status_code == 400, r.text
    assert r.json()["error"] == "Guild owner cannot be added as a member"

def test_add_member_unauthorized(client: Client, guild: dict, member_account: dict):
    r = client.post(f"/guilds/{guild['id']}/members/{member_account['id']}", json={})
    assert r.status_code == 401, r.text

def test_add_member_guild_not_found(owner: Client, member_account: dict):
    r = owner.post(f"/guilds/123345345/members/{member_account['id']}", json={})
    assert r.status_code == 404, r.text
    assert r.json()["error"] == "Guild not found"