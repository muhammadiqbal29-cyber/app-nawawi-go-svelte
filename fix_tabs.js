const fs = require('fs');

let content = fs.readFileSync('./frontend/src/routes/guru/konfigurasi/+page.svelte', 'utf8');

// 1. page-container CSS
content = content.replace(
`  .page-container {
    max-width: 900px;
    margin: 0 auto;
    padding-bottom: 3rem;
  }`,
`  .page-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    padding-bottom: 2rem;
  }`
);

// 2. tabs CSS
const oldTabsCss = `  /* Tabs */
  .tabs-header {
    display: inline-flex;
    background-color: var(--color-surface, #fff);
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 12px;
    padding: 0.35rem;
    box-shadow: var(--shadow-sm, 0 1px 2px rgba(0, 0, 0, 0.05));
    margin-bottom: 1.5rem;
  }
  .tab-btn {
    padding: 0.6rem 1.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    border: none;
    background: transparent;
    color: var(--color-text-muted, #64748b);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    font-family: inherit;
  }
  .tab-btn.active {
    background-color: var(--color-primary, #4f46e5);
    color: white;
  }`;

const newTabsCss = `  /* Main Tabs */
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
    margin-bottom: -2px;
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
    color: var(--color-primary, #10b981);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }`;

content = content.replace(oldTabsCss, newTabsCss);

// 3. Main tabs and title layout
const oldLayout = `<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Konfigurasi Materi</h1>
    </div>
  </div>

  <!-- Pilih Sekolah via Layout -->
  {#if !selectedSekolah}
    <div class="empty-state">
      <Icon name="grid" size={48} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas.</p>
    </div>
  {:else}
    <!-- Tabs -->
    <div class="tabs">
      {#if hasTahfiz}
        <button
          class="tab-btn"
          class:active={activeTab === 'tahfiz'}
          onclick={() => (activeTab = 'tahfiz')}
        >
          Materi Tahfiz
        </button>
      {/if}
      {#if hasDiniyah}
        <button
          class="tab-btn"
          class:active={activeTab === 'diniyah'}
          onclick={() => (activeTab = 'diniyah')}
        >
          Mapel Diniyah
        </button>
      {/if}
      {#if hasKeseharian}
        <button
          class="tab-btn"
          class:active={activeTab === 'keseharian'}
          onclick={() => (activeTab = 'keseharian')}
        >
          Keseharian
        </button>
      {/if}
      {#if hasTahfiz || hasDiniyah || hasKeseharian}
        <button
          class="tab-btn"
          class:active={activeTab === 'kop_logo'}
          onclick={() => (activeTab = 'kop_logo')}
        >
          Kop & Logo Rapor
        </button>
      {/if}
    </div>`;

const newLayout = `<div class="page-container">
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
    </div>`;

content = content.replace(oldLayout, newLayout);

// 4. Sub tabs Tahfiz
const oldTahfizTabs = `        <div
          style="display: flex; gap: 0.5rem; margin-bottom: 1.5rem; background: #f1f5f9; border: 1px solid #cbd5e1; padding: 0.35rem; border-radius: 999px; max-width: 750px;"
        >
          <button
            style="flex: 1; border-radius: 999px; padding: 0.5rem 1rem; border: none; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s; {tahfizActiveTab ===
            'materi'
              ? 'background: #10b981; color: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1);'
              : 'background: transparent; color: #64748b;'}"
            onclick={() => (tahfizActiveTab = 'materi')}
          >
            Daftar Materi Hafalan
          </button>
          <button
            style="flex: 1; border-radius: 999px; padding: 0.5rem 1rem; border: none; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s; {tahfizActiveTab ===
            'kelulusan'
              ? 'background: #10b981; color: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1);'
              : 'background: transparent; color: #64748b;'}"
            onclick={() => (tahfizActiveTab = 'kelulusan')}
          >
            Konfigurasi Nilai Kelulusan
          </button>
          <button
            style="flex: 1; border-radius: 999px; padding: 0.5rem 1rem; border: none; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s; {tahfizActiveTab ===
            'kriteria'
              ? 'background: #10b981; color: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1);'
              : 'background: transparent; color: #64748b;'}"
            onclick={() => (tahfizActiveTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>`;
const newTahfizTabs = `        <div class="card-tabs">
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
        </div>`;
content = content.replace(oldTahfizTabs, newTahfizTabs);

// 5. Sub tabs Diniyah
const oldDiniyahTabs = `        <div
          style="display: flex; gap: 0.5rem; margin-bottom: 1.5rem; background: #f1f5f9; border: 1px solid #cbd5e1; padding: 0.35rem; border-radius: 999px; max-width: 750px;"
        >
          <button
            style="flex: 1; border-radius: 999px; padding: 0.5rem 1rem; border: none; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s; {diniyahActiveTab ===
            'mapel'
              ? 'background: #10b981; color: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1);'
              : 'background: transparent; color: #64748b;'}"
            onclick={() => (diniyahActiveTab = 'mapel')}
          >
            Daftar Mapel
          </button>
          <button
            style="flex: 1; border-radius: 999px; padding: 0.5rem 1rem; border: none; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s; {diniyahActiveTab ===
            'kriteria'
              ? 'background: #10b981; color: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1);'
              : 'background: transparent; color: #64748b;'}"
            onclick={() => (diniyahActiveTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>`;
const newDiniyahTabs = `        <div class="card-tabs">
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
        </div>`;
content = content.replace(oldDiniyahTabs, newDiniyahTabs);

// 6. Sub tabs Keseharian
const oldKeseharianTabs = `        <div
          class="tabs sub-tabs"
          style="margin-bottom: 1rem; border-bottom: 1px solid var(--color-border); padding-bottom: 0;"
        >
          <button
            class="tab-btn"
            class:active={activeSubTab === 'keseharian'}
            onclick={() => (activeSubTab = 'keseharian')}>Indikator Keseharian</button
          >
          <button
            class="tab-btn"
            class:active={activeSubTab === 'pelanggaran'}
            onclick={() => (activeSubTab = 'pelanggaran')}>Pelanggaran</button
          >
          <button
            class="tab-btn"
            class:active={activeSubTab === 'iqob'}
            onclick={() => (activeSubTab = 'iqob')}>Iqob</button
          >
          <button
            class="tab-btn"
            class:active={activeSubTab === 'kriteria'}
            onclick={() => (activeSubTab = 'kriteria')}>Kriteria Penilaian</button
          >
        </div>`;
const newKeseharianTabs = `        <div class="card-tabs">
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
        </div>`;
content = content.replace(oldKeseharianTabs, newKeseharianTabs);

// Add mobile CSS before </style>
const mobileCss = `
  @media (max-width: 768px) {
    .tabs-header {
      border-bottom: none;
      margin-bottom: 0.5rem;
      padding-bottom: 0.5rem;
      gap: 10px;
    }
    
    .tab-link {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 7px;
      padding: 12px 6px 10px;
      border-radius: 18px;
      background: var(--color-surface, #fff);
      border: 1px solid var(--color-border, #e2e8f0);
      margin-bottom: 0;
    }

    .tab-link:hover {
      background: var(--color-surface, #fff);
      border-radius: 18px;
    }

    .tab-link.active {
      background: var(--color-primary, #22c55e);
      border-color: var(--color-primary, #22c55e);
      color: #fff;
      transform: translateY(-2px);
      box-shadow: 0 10px 30px -12px rgba(34, 197, 94, 0.4);
    }
    
    .icon-wrapper {
      width: 34px;
      height: 34px;
      border-radius: 11px;
      background: #e6f7ec;
      color: var(--color-primary, #22c55e);
      transition: all 0.2s;
    }
    
    .tab-link.active .icon-wrapper {
      background: rgba(255,255,255,0.25);
      color: #fff;
    }

    .tab-link span {
      font-size: 11.5px;
      font-weight: 700;
    }

    .card-tabs {
      display: flex;
      width: 100%;
      overflow-x: auto;
      scrollbar-width: none;
      -ms-overflow-style: none;
      background: transparent;
      border: none;
      padding: 0;
      gap: 8px;
      margin-bottom: 1rem;
    }
    .card-tabs::-webkit-scrollbar {
      display: none;
    }
    
    .card-tab-btn {
      padding: 9px 14px;
      border-radius: 12px;
      font-size: 12.5px;
      background: var(--color-surface, #fff);
      border: 1px solid var(--color-border, #e2e8f0);
      box-shadow: none;
    }
    
    .card-tab-btn.active {
      background: var(--color-text-main, #1e293b);
      border-color: var(--color-text-main, #1e293b);
      color: #fff;
      box-shadow: none;
    }
  }
`;

content = content.replace('</style>', mobileCss + '\n</style>');

fs.writeFileSync('./frontend/src/routes/guru/konfigurasi/+page.svelte', content, 'utf8');
