<script>
  import { guruState } from '$lib/guruState.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { get } from 'svelte/store';

  function selectSekolah(sekolah) {
    guruState.activeSekolahId = sekolah.id;
    guruState.activeSekolahName = sekolah.nama_sekolah;
    guruState.showJenjangModal = false;
    if (typeof window !== 'undefined') {
      localStorage.setItem('siakad_active_sekolah_id', sekolah.id);
      localStorage.setItem('siakad_active_sekolah_name', sekolah.nama_sekolah);
    }

    // Jika user sedang berada di halaman spesifik, redirect ke dashboard untuk mereset state dan versi (V1/V2)
    const currentPage = get(page);
    const path = currentPage.url.pathname;
    const urlSekolahId = currentPage.url.searchParams.get('sekolah_id');
    
    if (path.startsWith('/guru/input-nilai') || path.startsWith('/guru/rapor') || (urlSekolahId && urlSekolahId !== String(sekolah.id))) {
      goto('/guru');
    }
  }
</script>

{#if guruState.showJenjangModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-backdrop" onclick={(e) => { if (e.target === e.currentTarget) guruState.showJenjangModal = false; }}>
    <div class="modal-card">
      <div class="modal-header">
        <div class="icon-circle">
          <Icon name="grid" size={24} />
        </div>
        <h2>Pilih Jenjang Sekolah</h2>
        <p>Silakan pilih jenjang sekolah yang ingin Anda kelola saat ini.</p>
      </div>

      <div class="modal-body">
        {#if guruState.sekolahList.length === 0}
          <div class="empty-state">
            <p>Memuat jenjang...</p>
          </div>
        {:else}
          <div class="jenjang-grid">
            {#each guruState.sekolahList as sekolah (sekolah.id)}
              <button
                class="jenjang-card {guruState.activeSekolahId === sekolah.id ? 'active' : ''}"
                onclick={() => selectSekolah(sekolah)}
              >
                <div class="jenjang-info">
                  <span class="jenjang-name">{sekolah.nama_sekolah}</span>
                  <span class="jenjang-type">{sekolah.tingkatan}</span>
                </div>
                {#if guruState.activeSekolahId === sekolah.id}
                  <div class="check-icon">
                    <Icon name="check" size={20} />
                  </div>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </div>

      {#if guruState.activeSekolahId}
        <div class="modal-footer">
          <button class="btn-tutup" onclick={() => (guruState.showJenjangModal = false)}>
            Batal
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(15, 23, 42, 0.6);
    backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }
  .modal-card {
    background: white;
    border-radius: 1.5rem;
    width: 100%;
    max-width: 500px;
    box-shadow:
      0 20px 25px -5px rgba(0, 0, 0, 0.1),
      0 10px 10px -5px rgba(0, 0, 0, 0.04);
    overflow: hidden;
    animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }
  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  .modal-header {
    padding: 2rem 2rem 1.5rem;
    text-align: center;
    border-bottom: 1px solid #f1f5f9;
  }
  .icon-circle {
    width: 56px;
    height: 56px;
    background: #eef2ff;
    color: #4f46e5;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 1rem;
  }
  .modal-header h2 {
    margin: 0 0 0.5rem;
    font-size: 1.25rem;
    color: #0f172a;
  }
  .modal-header p {
    margin: 0;
    color: #64748b;
    font-size: 0.95rem;
  }
  .modal-body {
    padding: 1.5rem 2rem;
    background: #f8fafc;
  }
  .jenjang-grid {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .jenjang-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 1rem 1.25rem;
    background: white;
    border: 1px solid #cbd5e1;
    border-radius: 1rem;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s;
  }
  .jenjang-card:hover {
    border-color: #94a3b8;
    background: #f8fafc;
  }
  .jenjang-card.active {
    border-color: #4f46e5;
    background: #eef2ff;
    box-shadow: 0 4px 6px -1px rgba(79, 70, 229, 0.1);
  }
  .jenjang-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .jenjang-name {
    font-weight: 600;
    color: #1e293b;
  }
  .jenjang-type {
    font-size: 0.85rem;
    color: #64748b;
  }
  .check-icon {
    color: #4f46e5;
  }
  .empty-state {
    text-align: center;
    color: #64748b;
    padding: 2rem 0;
  }
  .modal-footer {
    padding: 1rem 2rem;
    background: white;
    border-top: 1px solid #f1f5f9;
    display: flex;
    justify-content: flex-end;
  }
  .btn-tutup {
    padding: 0.5rem 1rem;
    background: transparent;
    border: none;
    color: #64748b;
    font-weight: 500;
    cursor: pointer;
    border-radius: 0.5rem;
  }
  .btn-tutup:hover {
    background: #f1f5f9;
    color: #334155;
  }
</style>
