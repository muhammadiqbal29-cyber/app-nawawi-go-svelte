<script>
  import InputNilai from '$lib/components/guru/InputNilai.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { guruState } from '$lib/guruState.svelte.js';
  import { goto } from '$app/navigation';

  let hasAccess = $derived(
    guruState.activeSekolahId &&
      guruState.aksesMap[guruState.activeSekolahId] &&
      (guruState.aksesMap[guruState.activeSekolahId].includes('input_tahfiz_v2') || guruState.aksesMap[guruState.activeSekolahId].includes('paket_alquran'))
  );

  $effect(() => {
    if (guruState.activeSekolahId && !hasAccess) {
      const akses = guruState.aksesMap[guruState.activeSekolahId] || [];
      if (akses.includes('input_tahfiz')) {
        goto('/guru/input-nilai/tahfiz', { replaceState: true });
      }
    }
  });
</script>

{#if !guruState.activeSekolahId}
  <div class="empty-state">
    <Icon name="grid" size={48} />
    <h3>Pilih Jenjang Sekolah</h3>
    <p>Silakan klik tombol Pilih Jenjang terlebih dahulu.</p>
  </div>
{:else if !hasAccess}
  <div class="empty-state">
    <Icon name="alertCircle" size={48} />
    <h3>Akses Ditolak</h3>
    <p>
      Anda tidak memiliki hak akses <b>Input Nilai</b> untuk jenjang {guruState.activeSekolahName}.
    </p>
  </div>
{:else}
  {#key guruState.activeSekolahId}
    <InputNilai sekolahId={guruState.activeSekolahId} isV2={true} />
  {/key}
{/if}

<style>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
    color: #64748b;
    background: white;
    border-radius: 1rem;
    border: 1px solid #e2e8f0;
  }
  .empty-state h3 {
    margin: 0;
    color: #0f172a;
  }

  /* --- Mobile Card Layout Injected --- */
  @media (max-width: 768px) {
    .table-container {
      background: transparent !important;
      border: none !important;
      box-shadow: none !important;
      padding: 0 !important;
    }
    .data-table, .data-table tbody, .data-table tr, .data-table td {
      display: block;
      width: 100%;
    }
    .data-table thead {
      display: none;
    }
    .data-table tr {
      margin-bottom: 1.25rem;
      border: 1px solid #cbd5e1;
      border-radius: 12px;
      background: #fff;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
      overflow: hidden;
    }
    .data-table td {
      padding: 1rem;
      border-bottom: 1px solid #f1f5f9;
      text-align: left !important;
    }
    .data-table td:last-child {
      border-bottom: none;
      background-color: #f8fafc;
    }
    .siswa-nama {
      font-size: 1.1rem;
      color: #0f172a;
    }
    .page-header-actions {
      flex-direction: column;
      align-items: stretch;
    }
    .action-buttons {
      width: 100%;
      flex-direction: column;
      margin-top: 1rem;
    }
    .btn-action-primary, .btn-action-secondary {
      width: 100%;
      justify-content: center;
    }
    .search-bar {
      width: 100%;
    }
    textarea, input[type="number"], .form-input {
      width: 100% !important;
      min-width: 0 !important;
    }
    .student-row {
      flex-direction: column;
      align-items: stretch;
    }
    .student-info-group {
      margin-bottom: 0.5rem;
    }
  }
</style>

