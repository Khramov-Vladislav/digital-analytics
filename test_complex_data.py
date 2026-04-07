#!/usr/bin/env python3
"""Test complex/non-standard resume data handling"""

import urllib.request
import json
import io

# Test data with complex/non-standard values
test_data = {
    'fullname': 'Александр-Сергей Петрович Иванов-Смирнов',
    'phone': '+7 (999) 123-45-67',
    'email': 'alexander.ivanov-smirnov+test2026@mail-example.co.uk',
    'city': 'Санкт-Петербург',
    'summary': 'Full-stack разработчик с 8+ годами опыта. Специализируюсь на высоконагруженных системах, микросервисной архитектуре и оптимизации производительности. Работал с международными командами и участвовал в разработке продуктов с аудиторией более 1 млн пользователей.',
    'template': 'creative',
    'skills': 'JavaScript, TypeScript, React, Node.js, Express, NestJS, PostgreSQL, MongoDB, Docker, Kubernetes, AWS, Git, CI/CD, GraphQL, REST API, HTML5, CSS3, Webpack',
    'extra': 'Английский язык: C1 (Advanced)\nГотов к релокации и командировкам\nУчаствовал в open-source проектах (GitHub: github.com/example)\nХобби: бег, шахматы, фотография\nОпыт выступлений на конференциях',
}

# Education entries
edu_data = [
    {
        'institution': 'Санкт-Петербургский государственный университет, факультет прикладной математики и процессов управления',
        'degree': 'Бакалавр',
        'start_year': '2012',
        'end_year': '2016'
    },
    {
        'institution': 'Высшая школа экономики',
        'degree': 'Магистр программной инженерии',
        'start_year': '2016',
        'end_year': '2018'
    },
    {
        'institution': 'Skillbox - Advanced React & Architecture',
        'degree': 'Сертификат',
        'start_year': '2020',
        'end_year': '2021'
    }
]

# Experience entries
exp_data = [
    {
        'company': 'ООО "ТехноСофт"',
        'position': 'Senior Full-Stack Developer',
        'period': '2021-н.в.',
        'description': 'Разработал и внедрил микросервисную архитектуру (Node.js, Docker, Kubernetes). Оптимизировал API, снизив время ответа на 35%. Руководил командой из 5 разработчиков'
    },
    {
        'company': 'ЗАО "ИнфоСистемы"',
        'position': 'Middle Developer',
        'period': '2018-2021',
        'description': 'Разработка frontend на React + TypeScript. Интеграция REST и GraphQL API. Покрытие кода тестами (Jest, Cypress)'
    }
]

print("=" * 60)
print("Testing Complex Resume Data")
print("=" * 60)

# Build multipart form data
boundary = '----WebKitFormBoundary' + ''.join([format(ord(c), 'x') for c in 'AbCdEfGhIjKl'])

body = io.BytesIO()

# Add basic fields
for key, value in test_data.items():
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="{key}"\r\n\r\n'.encode())
    body.write(f'{value}\r\n'.encode())

# Add education entries
for i, edu in enumerate(edu_data):
    for key, value in edu.items():
        body.write(f'--{boundary}\r\n'.encode())
        body.write(f'Content-Disposition: form-data; name="edu_{key}"\r\n\r\n'.encode())
        body.write(f'{value}\r\n'.encode())

# Add experience entries
for i, exp in enumerate(exp_data):
    for key, value in exp.items():
        body.write(f'--{boundary}\r\n'.encode())
        body.write(f'Content-Disposition: form-data; name="exp_{key}"\r\n\r\n'.encode())
        body.write(f'{value}\r\n'.encode())

body.write(f'--{boundary}--\r\n'.encode())

# Test 1: Generate PDF
print("\n1️⃣  Test: Generate PDF with complex data")
print("-" * 60)

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
        content_type = response.headers.get('Content-Type')
        
        if status == 200 and content_type == 'application/pdf':
            print(f"✅ PDF Generated successfully!")
            print(f"   Status: {status}")
            print(f"   Content-Type: {content_type}")
            print(f"   PDF Size: {len(pdf_data)} bytes")
            
            # Save PDF for manual inspection
            with open('/tmp/complex_resume.pdf', 'wb') as f:
                f.write(pdf_data)
            print(f"   Saved to: /tmp/complex_resume.pdf")
        else:
            print(f"❌ Unexpected response!")
            print(f"   Status: {status}")
            print(f"   Content-Type: {content_type}")
except urllib.error.HTTPError as e:
    error_msg = e.read().decode('utf-8')
    print(f"❌ HTTP Error {e.code}: {error_msg}")
except Exception as e:
    print(f"❌ Error: {e}")

# Test 2: Preview HTML
print("\n2️⃣  Test: Generate preview with complex data")
print("-" * 60)

# Reset body for preview request
body = io.BytesIO()
for key, value in test_data.items():
    body.write(f'--{boundary}\r\n'.encode())
    body.write(f'Content-Disposition: form-data; name="{key}"\r\n\r\n'.encode())
    body.write(f'{value}\r\n'.encode())

for i, edu in enumerate(edu_data):
    for key, value in edu.items():
        body.write(f'--{boundary}\r\n'.encode())
        body.write(f'Content-Disposition: form-data; name="edu_{key}"\r\n\r\n'.encode())
        body.write(f'{value}\r\n'.encode())

for i, exp in enumerate(exp_data):
    for key, value in exp.items():
        body.write(f'--{boundary}\r\n'.encode())
        body.write(f'Content-Disposition: form-data; name="exp_{key}"\r\n\r\n'.encode())
        body.write(f'{value}\r\n'.encode())

body.write(f'--{boundary}--\r\n'.encode())

try:
    req = urllib.request.Request(
        'http://localhost:8080/preview',
        data=body.getvalue(),
        method='POST'
    )
    req.add_header('Content-Type', f'multipart/form-data; boundary={boundary}')
    
    with urllib.request.urlopen(req) as response:
        html_data = response.read().decode('utf-8')
        status = response.status
        
        if status == 200:
            print(f"✅ Preview generated successfully!")
            print(f"   Status: {status}")
            print(f"   HTML Size: {len(html_data)} bytes")
            
            # Check for key data presence in HTML
            checks = [
                ('ФИО (Иванов)', 'Иванов'),
                ('Email (domain)', 'mail-example.co.uk'),
                ('Город', 'Санкт-Петербург'),
                ('Телефон (код)', '999'),
                ('Компания', 'ТехноСофт'),
                ('Должность', 'Senior Full-Stack'),
                ('Образование', 'Магистр'),
                ('Навык JavaScript', 'JavaScript'),
                ('Навык Docker', 'Docker'),
                ('Навык Kubernetes', 'Kubernetes'),
                ('Краткая инфо', 'Full-stack'),
            ]
            
            print("\n   Data presence checks:")
            for label, search_str in checks:
                if search_str in html_data:
                    print(f"   ✅ {label}: found '{search_str}'")
                else:
                    print(f"   ❌ {label}: NOT found '{search_str}'")
            
            # Save HTML for manual inspection
            with open('/tmp/complex_resume.html', 'w') as f:
                f.write(html_data)
            print(f"\n   Saved preview to: /tmp/complex_resume.html")
        else:
            print(f"❌ Unexpected response status: {status}")
            
except urllib.error.HTTPError as e:
    error_msg = e.read().decode('utf-8')
    print(f"❌ HTTP Error {e.code}: {error_msg}")
except Exception as e:
    print(f"❌ Error: {e}")

print("\n" + "=" * 60)
print("Testing complete! ✅")
print("=" * 60)
