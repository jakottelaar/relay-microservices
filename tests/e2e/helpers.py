import httpx
from PIL import Image
import io


def sign_up(client: httpx.Client, user: dict):
    client.post("/auth/sign-up", json=user)


def sign_in(client: httpx.Client, email: str, password: str) -> tuple[httpx.Client, dict]:
    r = client.post("/auth/sign-in", json={"email": email, "password": password})
    assert r.status_code == 200, r.text
    data = r.json()
    auth = httpx.Client(
        base_url=client.base_url,
        headers={"Authorization": f"Bearer {data['access_token']}"},
        timeout=10,
    )
    return auth, data["account"]


def create_guild(client: httpx.Client, name: str, description: str = None) -> dict:
    payload = {"name": name}
    if description:
        payload["description"] = description
    r = client.post("/guilds", data=payload)
    assert r.status_code == 201, r.text
    return r.json()


def create_channel(client: httpx.Client, guild_id: str, name: str, type: int = 1) -> dict:
    r = client.post(f"/guilds/{guild_id}/channels", json={"name": name, "type": type})
    assert r.status_code == 201, r.text
    return r.json()


def create_test_png() -> bytes:
    img = Image.new("RGB", (1, 1), color=(255, 0, 0))
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()