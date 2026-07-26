import re

with open('./frontend/src/routes/guru/konfigurasi/+page.svelte', 'r') as f:
    content = f.read()

# 1. Replace CSS page-container and tabs-header
css_page_container = """  .page-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    padding-bottom: 2rem;
  }"""

content = re.sub(r'  \.page-container \{\s*max-width: 900px;\s*margin: 0 auto;\s*padding-bottom: 3rem;\s*\}', css_page_container, content)

css_tabs_header_replacement = """  /* Main Tabs */
  .tabs-header {
    display: flex;
    gap: 0.5rem;
    border-bottom: 2px solid var(--color-border, #e2e8f0);
    margin-bottom: 1.5rem;
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }

  .tabs-header::-webkit-scrollbar {
    display: none;
  }

  .icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tab-link {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.25rem;
    color: var(--color-text-muted, #64748b);
    font-weight: 700;
    font-size: 0.95rem;
    text-decoration: none;
    border-bottom: 3px solid transparent;
    margin-bottom: -2px; /* overlap the border */
    transition: all 0.2s;
    white-space: nowrap;
    background: transparent;
    border-top: none;
    border-left: none;
    border-right: none;
    cursor: pointer;
  }

  .tab-link:hover {
    color: var(--color-primary, #4f46e5);
    background-color: var(--color-background, #f8fafc);
    border-radius: 8px 8px 0 0;
  }

  .tab-link.active {
    color: var(--color-primary, #4f46e5);
    border-bottom-color: var(--color-primary, #4f46e5);
  }

  /* Sub Tabs */
  .card-tabs {
    display: inline-flex;
    align-self: flex-start;
    background-color: #f1f5f9;
    border-radius: 999px;
    padding: 0.35rem;
    margin-bottom: 1.5rem;
    border: 1px solid #e2e8f0;
  }
  .card-tab-btn {
    padding: 0.6rem 1.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    border: none;
    background: transparent;
    color: var(--color-text-muted, #64748b);
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.3s ease;
    font-family: inherit;
    white-space: nowrap;
  }
  .card-tab-btn.active {
    background-color: white;
    color: var(--color-primary, #10b981); /* Green for konfigurasi */
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }"""

# Remove old .tabs-header and .tab-btn
old_css_regex = r'  /\* Tabs \*/\s*\.tabs-header \{[\s\S]*?\}\s*\.tab-btn \{[\s\S]*?\}\s*\.tab-btn\.active \{[\s\S]*?\}'
content = re.sub(old_css_regex, css_tabs_header_replacement, content)

# 2. Update Main Tabs HTML
main_tabs_html = """    <!-- Tabs -->
    <div class="tabs-header">
      {#if hasTahfiz}
        <button
          class="tab-link"
          class:active={activeTab === 'tahfiz'}
          onclick={() => (activeTab = 'tahfiz')}
        >
          <div class="icon-wrapper"><Icon name="bookOpen" size={18} /></div>
          <span>Materi Tahfiz</span>
        </button>
      {/if}
      {#if hasDiniyah}
        <button
          class="tab-link"
          class:active={activeTab === 'diniyah'}
          onclick={() => (activeTab = 'diniyah')}
        >
          <div class="icon-wrapper"><Icon name="book" size={18} /></div>
          <span>Mapel Diniyah</span>
        </button>
      {/if}
      {#if hasKeseharian}
        <button
          class="tab-link"
          class:active={activeTab === 'keseharian'}
          onclick={() => (activeTab = 'keseharian')}
        >
          <div class="icon-wrapper"><Icon name="smile" size={18} /></div>
          <span>Keseharian</span>
        </button>
      {/if}
      {#if hasTahfiz || hasDiniyah || hasKeseharian}
        <button
          class="tab-link"
          class:active={activeTab === 'kop_logo'}
          onclick={() => (activeTab = 'kop_logo')}
        >
          <div class="icon-wrapper"><Icon name="image" size={18} /></div>
          <span>Kop & Logo Rapor</span>
        </button>
      {/if}
    </div>"""

content = re.sub(r'    <!-- Tabs -->\s*<div class="tabs">[\s\S]*?</div>\s*(?={#if !hasTahfiz)', main_tabs_html + '\n\n    ', content)

# 3. Swap main tabs and title
page_container_html = """<div class="page-container">
  <!-- Pilih Sekolah via Layout -->
  {#if !selectedSekolah}
    <div class="empty-state">
      <Icon name="grid" size={48} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas.</p>
    </div>
  {:else}
    <!-- Tabs -->
    <div class="tabs-header">
      {#if hasTahfiz}
        <button
          class="tab-link"
          class:active={activeTab === 'tahfiz'}
          onclick={() => (activeTab = 'tahfiz')}
        >
          <div class="icon-wrapper"><Icon name="bookOpen" size={18} /></div>
          <span>Materi Tahfiz</span>
        </button>
      {/if}
      {#if hasDiniyah}
        <button
          class="tab-link"
          class:active={activeTab === 'diniyah'}
          onclick={() => (activeTab = 'diniyah')}
        >
          <div class="icon-wrapper"><Icon name="book" size={18} /></div>
          <span>Mapel Diniyah</span>
        </button>
      {/if}
      {#if hasKeseharian}
        <button
          class="tab-link"
          class:active={activeTab === 'keseharian'}
          onclick={() => (activeTab = 'keseharian')}
        >
          <div class="icon-wrapper"><Icon name="smile" size={18} /></div>
          <span>Keseharian</span>
        </button>
      {/if}
      {#if hasTahfiz || hasDiniyah || hasKeseharian}
        <button
          class="tab-link"
          class:active={activeTab === 'kop_logo'}
          onclick={() => (activeTab = 'kop_logo')}
        >
          <div class="icon-wrapper"><Icon name="image" size={18} /></div>
          <span>Kop & Logo Rapor</span>
        </button>
      {/if}
    </div>

    <div class="page-header">
      <div>
        <h1 class="page-title">Konfigurasi Materi</h1>
      </div>
    </div>"""

content = re.sub(r'<div class="page-container">\s*<div class="page-header">\s*<div>\s*<h1 class="page-title">Konfigurasi Materi</h1>\s*</div>\s*</div>\s*<!-- Pilih Sekolah via Layout -->\s*{#if !selectedSekolah}\s*<div class="empty-state">[\s\S]*?</div>\s*{:else}\s*<!-- Tabs -->\s*<div class="tabs-header">[\s\S]*?</div>', page_container_html, content)


# 4. Update sub tabs for Tahfiz
tahfiz_sub_tabs_replacement = """        <!-- Sub-tabs for Tahfiz -->
        <div class="card-tabs">
          <button
            class="card-tab-btn"
            class:active={tahfizActiveTab === 'materi'}
            onclick={() => (tahfizActiveTab = 'materi')}
          >
            Daftar Materi Hafalan
          </button>
          <button
            class="card-tab-btn"
            class:active={tahfizActiveTab === 'kelulusan'}
            onclick={() => (tahfizActiveTab = 'kelulusan')}
          >
            Konfigurasi Nilai Kelulusan
          </button>
          <button
            class="card-tab-btn"
            class:active={tahfizActiveTab === 'kriteria'}
            onclick={() => (tahfizActiveTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>"""

content = re.sub(r'        <!-- Sub-tabs for Tahfiz -->\s*<div\s*style="display: flex; gap: 0\.5rem; margin-bottom: 1\.5rem; background: #f1f5f9; border: 1px solid #cbd5e1; padding: 0\.35rem; border-radius: 999px; max-width: 750px;"\s*>\s*<button\s*style="flex: 1; border-radius: 999px; padding: 0\.5rem 1rem; border: none; font-weight: 600; font-size: 0\.9rem; cursor: pointer; transition: all 0\.2s; \{tahfizActiveTab ===\s*\'materi\'\s*\? \'background: #10b981; color: white; box-shadow: 0 1px 3px rgba\(0,0,0,0\.1\);\'\s*: \'background: transparent; color: #64748b;\'\}"\s*onclick=\{\(\) => \(tahfizActiveTab = \'materi\'\)\}\s*>\s*Daftar Materi Hafalan\s*</button>\s*<button\s*style="flex: 1; border-radius: 999px; padding: 0\.5rem 1rem; border: none; font-weight: 600; font-size: 0\.9rem; cursor: pointer; transition: all 0\.2s; \{tahfizActiveTab ===\s*\'kelulusan\'\s*\? \'background: #10b981; color: white; box-shadow: 0 1px 3px rgba\(0,0,0,0\.1\);\'\s*: \'background: transparent; color: #64748b;\'\}"\s*onclick=\{\(\) => \(tahfizActiveTab = \'kelulusan\'\)\}\s*>\s*Konfigurasi Nilai Kelulusan\s*</button>\s*<button\s*style="flex: 1; border-radius: 999px; padding: 0\.5rem 1rem; border: none; font-weight: 600; font-size: 0\.9rem; cursor: pointer; transition: all 0\.2s; \{tahfizActiveTab ===\s*\'kriteria\'\s*\? \'background: #10b981; color: white; box-shadow: 0 1px 3px rgba\(0,0,0,0\.1\);\'\s*: \'background: transparent; color: #64748b;\'\}"\s*onclick=\{\(\) => \(tahfizActiveTab = \'kriteria\'\)\}\s*>\s*Kriteria Penilaian\s*</button>\s*</div>', tahfiz_sub_tabs_replacement, content)

# 5. Update sub tabs for Diniyah
diniyah_sub_tabs_replacement = """        <!-- Sub-tabs for Diniyah -->
        <div class="card-tabs">
          <button
            class="card-tab-btn"
            class:active={diniyahActiveTab === 'mapel'}
            onclick={() => (diniyahActiveTab = 'mapel')}
          >
            Daftar Mapel
          </button>
          <button
            class="card-tab-btn"
            class:active={diniyahActiveTab === 'kriteria'}
            onclick={() => (diniyahActiveTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>"""

content = re.sub(r'        <!-- Sub-tabs for Diniyah -->\s*<div\s*style="display: flex; gap: 0\.5rem; margin-bottom: 1\.5rem; background: #f1f5f9; border: 1px solid #cbd5e1; padding: 0\.35rem; border-radius: 999px; max-width: 750px;"\s*>\s*<button\s*style="flex: 1; border-radius: 999px; padding: 0\.5rem 1rem; border: none; font-weight: 600; font-size: 0\.9rem; cursor: pointer; transition: all 0\.2s; \{diniyahActiveTab ===\s*\'mapel\'\s*\? \'background: #10b981; color: white; box-shadow: 0 1px 3px rgba\(0,0,0,0\.1\);\'\s*: \'background: transparent; color: #64748b;\'\}"\s*onclick=\{\(\) => \(diniyahActiveTab = \'mapel\'\)\}\s*>\s*Daftar Mapel\s*</button>\s*<button\s*style="flex: 1; border-radius: 999px; padding: 0\.5rem 1rem; border: none; font-weight: 600; font-size: 0\.9rem; cursor: pointer; transition: all 0\.2s; \{diniyahActiveTab ===\s*\'kriteria\'\s*\? \'background: #10b981; color: white; box-shadow: 0 1px 3px rgba\(0,0,0,0\.1\);\'\s*: \'background: transparent; color: #64748b;\'\}"\s*onclick=\{\(\) => \(diniyahActiveTab = \'kriteria\'\)\}\s*>\s*Kriteria Penilaian\s*</button>\s*</div>', diniyah_sub_tabs_replacement, content)

# 6. Update sub tabs for Keseharian
keseharian_sub_tabs_replacement = """        <div class="card-tabs">
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'keseharian'}
            onclick={() => (activeSubTab = 'keseharian')}>Indikator Keseharian</button>
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'pelanggaran'}
            onclick={() => (activeSubTab = 'pelanggaran')}>Pelanggaran</button>
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'iqob'}
            onclick={() => (activeSubTab = 'iqob')}>Iqob</button>
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'kriteria'}
            onclick={() => (activeSubTab = 'kriteria')}>Kriteria Penilaian</button>
        </div>"""
content = re.sub(r'        <div\s*class="tabs sub-tabs"\s*style="margin-bottom: 1rem; border-bottom: 1px solid var\(--color-border\); padding-bottom: 0;"\s*>\s*<button\s*class="tab-btn"\s*class:active=\{activeSubTab === \'keseharian\'\}\s*onclick=\{\(\) => \(activeSubTab = \'keseharian\'\)\}>Indikator Keseharian</button\s*>\s*<button\s*class="tab-btn"\s*class:active=\{activeSubTab === \'pelanggaran\'\}\s*onclick=\{\(\) => \(activeSubTab = \'pelanggaran\'\)\}>Pelanggaran</button\s*>\s*<button\s*class="tab-btn"\s*class:active=\{activeSubTab === \'iqob\'\}\s*onclick=\{\(\) => \(activeSubTab = \'iqob\'\)\}>Iqob</button\s*>\s*<button\s*class="tab-btn"\s*class:active=\{activeSubTab === \'kriteria\'\}\s*onclick=\{\(\) => \(activeSubTab = \'kriteria\'\)\}>Kriteria Penilaian</button\s*>\s*</div>', keseharian_sub_tabs_replacement, content)

with open('./frontend/src/routes/guru/konfigurasi/+page.svelte', 'w') as f:
    f.write(content)
