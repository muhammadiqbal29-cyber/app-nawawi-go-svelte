<script>
  /**
   * KonfirmasiMapelModal.svelte
   * Modal konfirmasi Svelte (bukan alert JS) yang muncul ketika guru
   * mencoba masuk ke mapel yang sudah diklaim oleh guru lain.
   *
   * Props:
   *  - isOpen: boolean
   *  - mapel: object — mapel yang ingin diakses
   *  - onLanjutkan: () => void — dipanggil jika guru memilih "Ya, Lanjutkan"
   *  - onBatalkan: () => void — dipanggil jika guru memilih "Batalkan"
   */
  import { fade, scale } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';

  let { isOpen = false, mapel = null, onLanjutkan, onBatalkan } = $props();

  function handleLanjutkan() {
    onLanjutkan?.();
  }

  function handleBatalkan() {
    onBatalkan?.();
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) {
      handleBatalkan();
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') handleBatalkan();
    if (e.key === 'Enter') handleLanjutkan();
  }
</script>

<svelte:window onkeydown={isOpen ? handleKeydown : null} />

{#if isOpen && mapel}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="modal-backdrop"
    onclick={handleBackdropClick}
    transition:fade={{ duration: 180 }}
  >
    <div
      class="modal-box"
      role="dialog"
      aria-modal="true"
      aria-labelledby="konfirmasi-title"
      transition:scale={{ duration: 220, easing: cubicOut, start: 0.92 }}
    >
      <!-- Icon peringatan -->
      <div class="modal-icon-wrap">
        <div class="modal-icon">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
            <line x1="12" y1="9" x2="12" y2="13"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
        </div>
      </div>

      <!-- Konten -->
      <div class="modal-content">
        <h3 id="konfirmasi-title" class="modal-title">Mapel Ini Sudah Ada Pengampunya</h3>
        <p class="modal-desc">
          Mapel <strong>{mapel.nama_mapel}</strong> sudah diklaim oleh
          <strong class="guru-nama">{mapel.guru_pengampu_nama || 'Guru Lain'}</strong>.
        </p>
        <p class="modal-desc-sub">
          Kamu yakin mau masuk dan input nilai di sini? Nilai yang kamu simpan
          akan menimpa data yang ada.
        </p>
      </div>

      <!-- Actions -->
      <div class="modal-actions">
        <button
          type="button"
          id="btn-batalkan-konfirmasi"
          class="btn-batal"
          onclick={handleBatalkan}
        >
          Batalkan
        </button>
        <button
          type="button"
          id="btn-lanjutkan-konfirmasi"
          class="btn-lanjut"
          onclick={handleLanjutkan}
        >
          Ya, Lanjutkan
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.5);
    backdrop-filter: blur(3px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    padding: 1rem;
  }

  .modal-box {
    background: #fff;
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(15, 23, 42, 0.2);
    max-width: 420px;
    width: 100%;
    padding: 2rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.25rem;
    text-align: center;
  }

  /* Icon */
  .modal-icon-wrap {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: #fff7ed;
    border: 2px solid #fed7aa;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: #ea580c;
  }

  /* Teks */
  .modal-title {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 800;
    color: #0f172a;
    line-height: 1.3;
  }

  .modal-desc {
    margin: 0;
    font-size: 0.9rem;
    color: #334155;
    line-height: 1.6;
  }

  .modal-desc-sub {
    margin: 0;
    font-size: 0.82rem;
    color: #64748b;
    line-height: 1.5;
  }

  .guru-nama {
    color: #ea580c;
  }

  .modal-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  /* Buttons */
  .modal-actions {
    display: flex;
    gap: 0.75rem;
    width: 100%;
    margin-top: 0.25rem;
  }

  .btn-batal,
  .btn-lanjut {
    flex: 1;
    padding: 0.65rem 1rem;
    border-radius: 8px;
    font-size: 0.9rem;
    font-weight: 700;
    cursor: pointer;
    border: none;
    transition: background 0.15s ease, transform 0.1s ease;
  }

  .btn-batal:active,
  .btn-lanjut:active {
    transform: scale(0.97);
  }

  .btn-batal {
    background: #f1f5f9;
    color: #475569;
    border: 1px solid #e2e8f0;
  }
  .btn-batal:hover {
    background: #e2e8f0;
  }

  .btn-lanjut {
    background: #ea580c;
    color: #fff;
  }
  .btn-lanjut:hover {
    background: #c2410c;
  }

  @media (max-width: 480px) {
    .modal-box {
      padding: 1.5rem 1.25rem;
    }
    .modal-actions {
      flex-direction: column-reverse;
    }
  }
</style>
