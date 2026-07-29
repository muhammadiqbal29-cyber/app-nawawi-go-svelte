import os

files = [
    "frontend/src/routes/(blank)/guru/input-nilai-v2/[siswaId]/+page.svelte",
    "frontend/src/routes/(blank)/guru/input-nilai-tasmi/[siswaId]/+page.svelte",
    "frontend/src/routes/(blank)/guru/input-nilai/[siswaId]/+page.svelte",
    "frontend/src/routes/(blank)/guru/input-nilai-tasmi-v2/[siswaId]/+page.svelte",
    "frontend/src/routes/guru/+page.svelte",
    "frontend/src/routes/guru/profil/+page.svelte",
    "frontend/src/routes/guru/konfigurasi/+page.svelte"
]

css_to_add = """
  .desktop-only { display: block; }
  .mobile-only { display: none !important; }
  @media (max-width: 768px) {
    .desktop-only { display: none !important; }
    .mobile-only { display: block !important; }
  }
"""

for f in files:
    if not os.path.exists(f):
        print(f"Not found: {f}")
        continue
    with open(f, 'r') as file:
        content = file.read()
    
    if '<style>' in content:
        content = content.replace('<style>', '<style>\n' + css_to_add, 1)
        with open(f, 'w') as file:
            file.write(content)
        print(f"Fixed {f}")
    else:
        # append <style> at the end
        with open(f, 'a') as file:
            file.write('\n<style>\n' + css_to_add + '\n</style>\n')
        print(f"Appended <style> to {f}")
