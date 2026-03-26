from httpx import Client
import pytest
from PIL import Image
import io

USER = {
    "email": "get-guild-tester@mail.com",
    "username": "get-guild-tester",
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

@pytest.fixture(scope="module", autouse=True)
def setup_test_guild(auth_client: Client, sign_in_user):
    png_data = create_test_png()
    files = {"icon": ("icon.png", png_data, "image/png")}

    r = auth_client.post("/guilds", data={"name": "Test Guild", "description": "A guild for testing"}, files=files)
    assert r.status_code == 201, r.text
    yield r.json()

def test_get_guild(auth_client: Client, setup_test_guild):
    guild_id = setup_test_guild["id"]
    r = auth_client.get(f"/guilds/{guild_id}")

    assert r.status_code == 200, r.text
    body = r.json()

    assert body["id"] == guild_id
    assert body["name"] == "Test Guild"
    assert body["description"] == "A guild for testing"
    assert body["owner_id"] == str(setup_test_guild["owner_id"])
    assert body["icon"] is not None

def test_get_guild_not_found(auth_client: Client):
    r = auth_client.get("/guilds/123453454564564564")

    assert r.status_code == 404, r.text
    body = r.json()
    assert body["error"] == "Guild not found"

def test_get_guild_unauthenticated(client: Client):
    r = client.get("/guilds/123453454564564564")

    assert r.status_code == 401, r.text