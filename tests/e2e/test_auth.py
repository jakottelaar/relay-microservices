import pytest
import httpx

def test_sign_up(client):
    r = client.post("/auth/sign-up", json={
            "email": "john-doe@mail.com",
            "password": "Secret1234",
            "username": "john-doe12"
            })
    assert r.status_code == 201, r.text