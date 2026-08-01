<script>
  /**
   * MapelCard.svelte
   * Card untuk setiap mata pelajaran diniyah.
   * Menampilkan status klaim dan memungkinkan guru mengklaim/melepas mapelnya.
   *
   * Props:
   *  - mapel: object MapelDiniyah dari API (dengan guru_pengampu_id & guru_pengampu_nama)
   *  - isSelected: boolean — apakah ini mapel yang sedang aktif di tabel input
   *  - currentGuruId: number | null — ID guru yang sedang login (null jika admin)
   *  - currentGuruNama: string — nama guru yang sedang login
   *  - onSelect: (mapel) => void — dipanggil saat card diklik untuk masuk ke tabel input
   *  - onKlaim: (mapelId, aksi: 'klaim'|'lepas') => Promise<void>
   */

  let {
    mapel,
    isSelected = false,
    currentGuruId = null,
    currentGuruNama = '',
    onSelect,
    onKlaim
  } = $props();

  let loadingKlaim = $state(false);

  // Status klaim
  let isMyMapel = $derived(
    currentGuruId !== null && mapel.guru_pengampu_id === currentGuruId
  );
  let isOthersMapel = $derived(
    mapel.guru_pengampu_id !== null &&
    mapel.guru_pengampu_id !== undefined &&
    !isMyMapel
  );
  let isUnclaimed = $derived(!mapel.guru_pengampu_id);

  async function handleKlaim(e, aksi) {
    e.stopPropagation(); // Jangan trigger onSelect
    loadingKlaim = true;
    try {
      await onKlaim(mapel.id, aksi);
    } finally {
      loadingKlaim = false;
    }
  }
</script>

<button
  type="button"
  class="mapel-card"
  class:selected={isSelected}
  class:my-mapel={isMyMapel}
  class:others-mapel={isOthersMapel}
  onclick={() => onSelect(mapel)}
  aria-pressed={isSelected}
>
  <!-- Ikon kategori -->
  <div class="card-icon" class:icon-tulis={mapel.kategori === 'tes_tulis'} class:icon-lisan={mapel.kategori === 'tes_lisan'}>
    {#if mapel.kategori === 'tes_lisan'}
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>
    {:else}
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
        <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
      </svg>
    {/if}
  </div>

  <!-- Nama mapel -->
  <div class="card-body">
    <span class="mapel-nama">{mapel.nama_mapel}</span>
    {#if mapel.nama_mapel_arab}
      <span class="mapel-arab">{mapel.nama_mapel_arab}</span>
    {/if}

    <!-- Badge status klaim -->
    {#if isMyMapel}
      <span class="badge badge-mine">✓ Mapelmu</span>
    {:else if isOthersMapel}
      <span class="badge badge-others">🔒 {mapel.guru_pengampu_nama || 'Guru Lain'}</span>
    {:else}
      <span class="badge badge-free">Belum diklaim</span>
    {/if}
  </div>

  <!-- Tombol klaim/lepas -->
  <div class="card-actions" onclick={(e) => e.stopPropagation()}>
    {#if isMyMapel}
      <button
        type="button"
        class="btn-klaim btn-lepas"
        onclick={(e) => handleKlaim(e, 'lepas')}
        disabled={loadingKlaim}
        title="Lepas klaim mapel ini"
      >
        {#if loadingKlaim}…{:else}Lepas{/if}
      </button>
    {:else if isUnclaimed}
      <button
        type="button"
        class="btn-klaim btn-claim"
        onclick={(e) => handleKlaim(e, 'klaim')}
        disabled={loadingKlaim}
        title="Klaim mapel ini sebagai pengampumu"
      >
        {#if loadingKlaim}…{:else}Klaim{/if}
      </button>
    {/if}
  </div>

  <!-- Indikator selected -->
  {#if isSelected}
    <div class="selected-dot" aria-hidden="true"></div>
  {/if}
</button>

<style>
  .mapel-card {
    position: relative;
    display: grid;
    grid-template-columns: 44px 1fr auto;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.85rem 1rem;
    border: 2px solid #e2e8f0;
    border-radius: 12px;
    background: #fff;
    cursor: pointer;
    text-align: left;
    font-family: inherit;
    transition: border-color 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
    overflow: hidden;
  }

  .mapel-card:hover {
    border-color: #a5b4fc;
    box-shadow: 0 4px 12px rgba(99, 102, 241, 0.1);
  }

  .mapel-card.selected {
    border-color: #6366f1;
    background: #eef2ff;
    box-shadow: 0 4px 16px rgba(99, 102, 241, 0.15);
  }

  .mapel-card.my-mapel {
    border-color: #86efac;
    background: #f0fdf4;
  }

  .mapel-card.my-mapel:hover {
    border-color: #4ade80;
    box-shadow: 0 4px 12px rgba(74, 222, 128, 0.15);
  }

  .mapel-card.my-mapel.selected {
    border-color: #22c55e;
    background: #dcfce7;
  }

  .mapel-card.others-mapel {
    border-color: #fed7aa;
    background: #fff7ed;
  }

  .mapel-card.others-mapel:hover {
    border-color: #fb923c;
    box-shadow: 0 4px 12px rgba(251, 146, 60, 0.15);
  }

  /* Icon area */
  .card-icon {
    width: 44px;
    height: 44px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .icon-tulis {
    background: #ede9fe;
    color: #7c3aed;
  }
  .icon-lisan {
    background: #fef3c7;
    color: #d97706;
  }

  /* Body */
  .card-body {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .mapel-nama {
    font-size: 0.9rem;
    font-weight: 700;
    color: #1e293b;
    line-height: 1.3;
  }

  .mapel-arab {
    font-size: 0.85rem;
    color: #64748b;
    font-family: 'Amiri', serif;
    direction: rtl;
  }

  /* Badges */
  .badge {
    display: inline-block;
    font-size: 0.72rem;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 999px;
    width: fit-content;
    margin-top: 2px;
  }

  .badge-mine {
    background: #dcfce7;
    color: #166534;
    border: 1px solid #86efac;
  }

  .badge-others {
    background: #fff7ed;
    color: #9a3412;
    border: 1px solid #fed7aa;
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .badge-free {
    background: #f1f5f9;
    color: #94a3b8;
    border: 1px solid #e2e8f0;
  }

  /* Action buttons */
  .card-actions {
    flex-shrink: 0;
  }

  .btn-klaim {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 4px 12px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    transition: background 0.15s ease, transform 0.1s ease;
    white-space: nowrap;
  }

  .btn-klaim:active {
    transform: scale(0.96);
  }

  .btn-klaim:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-claim {
    background: #6366f1;
    color: #fff;
  }
  .btn-claim:hover:not(:disabled) {
    background: #4f46e5;
  }

  .btn-lepas {
    background: #f1f5f9;
    color: #64748b;
    border: 1px solid #e2e8f0;
  }
  .btn-lepas:hover:not(:disabled) {
    background: #fee2e2;
    color: #dc2626;
    border-color: #fca5a5;
  }

  /* Selected indicator dot */
  .selected-dot {
    position: absolute;
    top: 10px;
    right: 10px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #6366f1;
  }
</style>
