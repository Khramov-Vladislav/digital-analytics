#!/usr/bin/env python3
"""Verify text wrapping in generated PDFs"""

import requests
import json
from pathlib import Path

def test_various_lengths():
    """Test names of different lengths to verify wrapping works"""
    
    test_cases = [
        ("Коротко Иванов", "short name (15 chars)"),
        ("Максимилиан Викторович Константинович", "medium name (37 chars)"),
        ("Максимилиан-Владимир Викторович Константинович-Емельянович-Шумахер-Абрамович", "very long (76 chars)"),
        ("Максимилиан-Владимир Викторович Константинович-Емельянович-Шумахер-Абрамович-Петрович", "extreme (88 chars)"),
    ]
    
    templates = ["classic", "modern", "professional", "creative"]
    
    print("=" * 70)
    print("TESTING TEXT WRAPPING WITH DIFFERENT NAME LENGTHS")
    print("=" * 70)
    
    for name, description in test_cases:
        print(f"\n📝 Test: {description}")
        print(f"   Name: {name}")
        print(f"   Characters: {len(name)}")
        
        for template in templates:
            data = {
                "fullname": name,
                "email": "test@example.com",
                "phone": "+79991234567",
                "city": "Москва",
                "summary": "Опытный специалист",
                "template": template,
                "skills": "",
                "extra": "",
            }
            
            try:
                response = requests.post(
                    "http://localhost:8080/generate",
                    files=[(k, (None, v)) for k, v in data.items()],
                    timeout=5
                )
                
                if response.status_code == 200:
                    size = len(response.content)
                    print(f"   ✅ {template:12} - {size:6} bytes")
                else:
                    error_msg = response.text
                    print(f"   ❌ {template:12} - Status {response.status_code}: {error_msg[:50]}")
            except Exception as e:
                print(f"   ❌ {template:12} - Error: {str(e)[:50]}")
    
    print("\n" + "=" * 70)
    print("Verification complete! Check generated PDFs for visual layout.")
    print("=" * 70)

if __name__ == "__main__":
    test_various_lengths()
