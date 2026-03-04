#!/usr/bin/env python3
import urllib.request
import urllib.parse
import json

BASE_URL = "http://localhost:8080"

def post_form(data):
    """Send form data and return status code and response text"""
    # Create multipart form data
    body = []
    for key, value in data.items():
        body.append(('--'+'-'*32).encode())
        body.append(f'Content-Disposition: form-data; name="{key}"'.encode())
        body.append(b'')
        body.append(value.encode())
    
    boundary = '-'*32
    body_bytes = b'\r\n'.join(body) + b'\r\n--' + boundary.encode() + b'--\r\n'
    
    try:
        req = urllib.request.Request(
            f"{BASE_URL}/generate",
            data=body_bytes,
            headers={
                'Content-Type': f'multipart/form-data; boundary={boundary}',
                'Content-Length': str(len(body_bytes))
            },
            method='POST'
        )
        
        with urllib.request.urlopen(req, timeout=10) as response:
            content = response.read()
            return response.status, response.headers.get('Content-Type', ''), len(content)
    except urllib.error.HTTPError as e:
        return e.code, '', e.read().decode()[:200]
    except Exception as e:
        return 0, '', str(e)[:200]

def test_validation():
    print("=" * 60)
    print("Testing Resume Generator Validation")
    print("=" * 60)
    
    # Test 1: Invalid phone (too short)
    print("\n1️⃣ Test: Invalid phone (only 3 digits)")
    print("-" * 60)
    data = {
        'fullname': 'Test User',
        'phone': '123',
        'email': 'test@test.com',
        'city': 'Test',
        'template': 'classic'
    }
    status, ctype, response = post_form(data)
    print(f"Status: {status}")
    if status != 200:
        print(f"✅ Validation error caught! Response: {response[:100]}...")
    else:
        print(f"Response: {response}")
    
    # Test 2: Invalid email
    print("\n2️⃣ Test: Invalid email format")
    print("-" * 60)
    data = {
        'fullname': 'Test User',
        'phone': '+79121234567',
        'email': 'invalidemail',
        'city': 'Test',
        'template': 'classic'
    }
    status, ctype, response = post_form(data)
    print(f"Status: {status}")
    if status != 200:
        print(f"✅ Validation error caught! Response: {response[:100]}...")
    else:
        print(f"Response: {response}")
    
    # Test 3: Valid data
    print("\n3️⃣ Test: Valid data (should generate PDF)")
    print("-" * 60)
    data = {
        'fullname': 'Иван Иванов',
        'phone': '+79121234567',
        'email': 'ivan@test.com',
        'city': 'Москва',
        'template': 'classic'
    }
    status, ctype, size = post_form(data)
    print(f"Status: {status}")
    print(f"Content-Type: {ctype}")
    if status == 200 and ctype == 'application/pdf':
        print(f"✅ PDF Generated! Size: {size} bytes")
    else:
        print(f"Response: {size}")
    
    # Test 4: Valid Russian phone formats  
    print("\n4️⃣ Test: Russian phone number (8xxx format)")
    print("-" * 60)
    data = {
        'fullname': 'Мария Петрова',
        'phone': '89121234567',  # Will be converted to +7
        'email': 'maria@example.com',
        'city': 'СПб',
        'template': 'modern'
    }
    status, ctype, size = post_form(data)
    print(f"Status: {status}")
    if status == 200 and 'pdf' in ctype.lower():
        print(f"✅ PDF Generated! Size: {size} bytes")
    else:
        print(f"Response: {size}")
    
    print("\n" + "=" * 60)
    print("All tests completed! ✅")
    print("=" * 60)

if __name__ == "__main__":
    test_validation()
