#!/usr/bin/env python3
"""Smoke tests for /generate: requires the app listening on BASE_URL."""

import json
import os
import urllib.error
import urllib.request

BASE_URL = os.environ.get("BASE_URL", "http://localhost:8080").rstrip("/")


def post_form(data):
    """Send multipart form data and return (status, content_type, body_or_message)."""
    body = []
    for key, value in data.items():
        body.append(("--" + "-" * 32).encode())
        body.append(f'Content-Disposition: form-data; name="{key}"'.encode())
        body.append(b"")
        body.append(value.encode())

    boundary = "-" * 32
    body_bytes = b"\r\n".join(body) + b"\r\n--" + boundary.encode() + b"--\r\n"

    try:
        req = urllib.request.Request(
            f"{BASE_URL}/generate",
            data=body_bytes,
            headers={
                "Content-Type": f"multipart/form-data; boundary={boundary}",
                "Content-Length": str(len(body_bytes)),
            },
            method="POST",
        )

        with urllib.request.urlopen(req, timeout=10) as response:
            content = response.read()
            return response.status, response.headers.get("Content-Type", ""), content
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            msg = raw.decode()
        except Exception:
            msg = str(raw)
        return e.code, e.headers.get("Content-Type", ""), msg
    except urllib.error.URLError as e:
        return None, "", f"URLError: {e.reason}"


def test_validation():
    print("=" * 60)
    print("Testing Resume Generator Validation")
    print(f"Endpoint: {BASE_URL}/generate")
    print("=" * 60)

    # Test 1: Invalid phone (too short)
    print("\n1️⃣ Test: Invalid phone (only 3 digits)")
    print("-" * 60)
    data = {
        "fullname": "Test User",
        "phone": "123",
        "email": "test@test.com",
        "city": "Москва",
        "template": "classic",
    }
    status, ctype, body = post_form(data)
    print(f"Status: {status}")
    if status is None:
        print(f"❌ Server unreachable: {body}")
    elif status == 400 and "json" in ctype.lower():
        try:
            err = json.loads(body)["error"]
            print(f"✅ Validation error (JSON): {err[:120]}")
        except Exception:
            print(f"✅ Validation error: {body[:120]}")
    else:
        print(f"❌ Expected HTTP 400 JSON, got status={status} body={repr(body)[:200]}")

    # Test 2: Invalid email
    print("\n2️⃣ Test: Invalid email format")
    print("-" * 60)
    data = {
        "fullname": "Test User",
        "phone": "+79121234567",
        "email": "invalidemail",
        "city": "Москва",
        "template": "classic",
    }
    status, ctype, body = post_form(data)
    print(f"Status: {status}")
    if status is None:
        print(f"❌ Server unreachable: {body}")
    elif status == 400:
        print(f"✅ Validation rejected bad email: {body[:120]}")
    else:
        print(f"❌ Expected HTTP 400, got {status}")

    # Test 3: Valid data
    print("\n3️⃣ Test: Valid data (should generate PDF)")
    print("-" * 60)
    data = {
        "fullname": "Иван Иванов",
        "phone": "+79121234567",
        "email": "ivan@test.com",
        "city": "Москва",
        "template": "classic",
    }
    status, ctype, body = post_form(data)
    print(f"Status: {status}")
    print(f"Content-Type: {ctype}")
    if status is None:
        print(f"❌ Server unreachable: {body}")
    elif status == 200 and ctype == "application/pdf":
        print(f"✅ PDF generated, {len(body)} bytes")
    else:
        print(f"❌ Expected PDF: {repr(body)[:200]}")

    # Test 4: Russian phone 8… and city from autocomplete list
    print("\n4️⃣ Test: Russian phone number (8xxx format), city Санкт-Петербург")
    print("-" * 60)
    data = {
        "fullname": "Мария Петрова",
        "phone": "89121234567",
        "email": "maria@example.com",
        "city": "Санкт-Петербург",
        "template": "modern",
    }
    status, ctype, body = post_form(data)
    print(f"Status: {status}")
    if status is None:
        print(f"❌ Server unreachable: {body}")
    elif status == 200 and "pdf" in ctype.lower():
        print(f"✅ PDF generated, {len(body)} bytes")
    else:
        print(f"❌ Expected PDF: {repr(body)[:200]}")

    print("\n" + "=" * 60)
    if status is None:
        print("Finish: start the server: make run  (or go run ./cmd)")
    else:
        print("All checks finished.")
    print("=" * 60)


if __name__ == "__main__":
    test_validation()
