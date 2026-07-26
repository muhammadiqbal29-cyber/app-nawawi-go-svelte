import re

with open('./frontend/src/routes/guru/konfigurasi/+page.svelte', 'r') as f:
    content = f.read()

# Find the last </style> and remove the new <style> tag
content = content.replace('</style>\n\n<style>', '')
content = content.replace('</style>\n<style>', '')

with open('./frontend/src/routes/guru/konfigurasi/+page.svelte', 'w') as f:
    f.write(content)
