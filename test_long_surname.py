#!/usr/bin/env python3
"""Test long surname handling in all templates"""

import urllib.request
import io

# Test data with VERY long surname
test_templates = [
    'classic',
    'professional', 
    'modern',
    'creative'
]

# Long surname: "Максимилиан-Владимир Викторович Константинович-Емельянович-Шумахер-Абрамович"
long_name = "Максимилиан-Владимир Викторович Константинович-Емельянович-Шумахер-Абрамович"

print("=" * 70)
print("Testing Long Surname Handling in All Templates")
print("=" * 70)
print(f"\nTest Surname: {long_name}")
print(f"Length: {len(long_name)} characters\n")

for template in test_templates:
    print(f"\n🧪 Testing template: {template}")
    print("-" * 70)
    
    # Build form data
    boundary = '----WebKitFormBoundary' + ''.join([format(ord(c), 'x') for c in 'AbCdEf'])
    body = io.BytesIO()
    
    fields = {
        'fullname': long_name,
        'phone': '+7 (999) 123-45-67',
        'email': 'test@example.com',
        'city': 'Москва',
        'summary': 'Опытный разработчик с 10 годами стажа',
        'template': template,
        'skills': 'Python, Go, JavaScript',
        'extra': 'Дополнительная информация',
    }
    
    for key, value in fields.items():
        body.write(f'--{boundary}\r\n'.encode())
        body.write(f'Content-Disposition: form-data; name="{key}"\r\n\r\n'.encode())
        body.write(f'{value}\r\n'.encode())
    
    # Add one education entry
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="edu_institution"\r\n\r\n'.encode())
    body.write(f'Московский государственный университет\r\n'.encode())
    
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="edu_degree"\r\n\r\n'.encode())
    body.write(f'Бакалавр\r\n'.encode())
    
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="edu_start_year"\r\n\r\n'.encode())
    body.write(f'2010\r\n'.encode())
    
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="edu_end_year"\r\n\r\n'.encode())
    body.write(f'2014\r\n'.encode())
    
    # Add one experience entry
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="exp_company"\r\n\r\n'.encode())
    body.write(f'ООО Компания\r\n'.encode())
    
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="exp_position"\r\n\r\n'.encode())
    body.write(f'Senior Developer\r\n'.encode())
    
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="exp_period"\r\n\r\n'.encode())
    body.write(f'2015-2024\r\n'.encode())
    
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="exp_desc"\r\n\r\n'.encode())
    body.write(f'Разработка приложений\r\n'.encode())
    
    body.write(f'--{boundary}--\r\n'.encode())
    
    # Test PDF generation
    try:
        req = urllib.request.Request(
            'http://localhost:8080/generate',
            data=body.getvalue(),
            method='POST'
        )
        req.add_header('Content-Type', f'multipart/form-data; boundary={boundary}')
        
        with urllib.request.urlopen(req) as response:
            pdf_data = response.read()
            status = response.status
            
            if status == 200:
                print(f"✅ PDF Generated: {len(pdf_data)} bytes")
                
                # Save PDF for inspection
                filename = f'/tmp/long_surname_{template}.pdf'
                with open(filename, 'wb') as f:
                    f.write(pdf_data)
                print(f"   Saved: {filename}")
            else:
                print(f"❌ Unexpected status: {status}")
    except urllib.error.HTTPError as e:
        error_msg = e.read().decode('utf-8')
        print(f"❌ HTTP Error {e.code}: {error_msg}")
    except Exception as e:
        print(f"❌ Error: {e}")

print("\n" + "=" * 70)
print("Testing complete!")
print("=" * 70)
print("\nGenerated PDFs can be inspected for proper text wrapping and layout")
