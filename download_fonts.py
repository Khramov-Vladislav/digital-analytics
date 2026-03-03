#!/usr/bin/env python3
import urllib.request
import ssl
import os

# Отключаем SSL проверку
ssl._create_default_https_context = ssl._create_unverified_context

os.chdir('fonts')

urls = {
    "DejaVuSans.ttf": "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSans.ttf",
    "DejaVuSans-Bold.ttf": "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSans-Bold.ttf"
}

for filename, url in urls.items():
    try:
        print(f"Скачиваю {filename}...")
        req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
        response = urllib.request.urlopen(req, timeout=10)
        with open(filename, 'wb') as f:
            data = response.read()
            f.write(data)
        size = os.path.getsize(filename)
        print(f"✓ {filename} ({size} bytes)")
    except Exception as e:
        print(f"✗ Ошибка при скачивании {filename}: {e}")
