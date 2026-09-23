<script>
  import Icon from '$lib/components/ui/Icon.svelte';
  import PresensiWidget from '$lib/components/guru/PresensiWidget.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

  let selectedSekolah = $derived(
    guruState.sekolahList.find((s) => s.id == guruState.activeSekolahId) || null
  );
</script>

<div class="beranda-guru desktop-only">
  {#if !selectedSekolah}
    <div class="empty-state">
      <Icon name="grid" size={48} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan klik tombol Pilih Jenjang terlebih dahulu untuk melihat dashboard.</p>
    </div>
  {:else}
    <!-- WIDGET PRESENSI (Langsung Tampil di Beranda Desktop) -->
    <div style="margin-bottom: 2rem;">
      <PresensiWidget />
    </div>
  {/if}
</div>

<!-- TAMPILAN MOBILE: BERANDA GURU -->
<div class="mobile-app-ui sheet-page mobile-only" style="padding-top: 0 !important; margin-bottom: 80px;">
  <div class="mobile-app-header">
    <h1 class="page-title">Beranda Guru</h1>
    <div class="student-row" style="margin-top: 8px;">
      <div class="student-avatar" style="background: linear-gradient(135deg, #10b981, #047857);">
        {#if selectedSekolah}
          {selectedSekolah.nama_sekolah.charAt(0)}
        {:else}
          -
        {/if}
      </div>
      <div class="student-sub">
        {#if selectedSekolah}
          <b>{selectedSekolah.nama_sekolah}</b><br>Tahun Ajaran Aktif
        {:else}
          Silakan klik tombol Pilih Jenjang terlebih dahulu
        {/if}
      </div>
    </div>
  </div>

  {#if selectedSekolah}
    <!-- WIDGET PRESENSI (Langsung Tampil di Beranda) -->
    <div style="margin: 0 1rem;">
      <PresensiWidget />
    </div>
  {/if}
</div>

<style>
  .desktop-only { display: block; }
  .mobile-only { display: none !important; }
  @media (max-width: 768px) {
    .desktop-only { display: none !important; }
    .mobile-only { display: block !important; }
  }

  .beranda-guru {
    max-width: 1200px;
    margin: 0 auto;
    font-family: 'Inter', sans-serif;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #64748b;
    background: white;
    border-radius: 1rem;
    border: 1px dashed #cbd5e1;
  }

  .empty-state h3 {
    margin: 0;
    color: #0f172a;
  }

  .mobile-app-ui {
    background: #f8fafc;
    min-height: 100vh;
  }
  
  .mobile-app-header {
    background: #ffffff;
    padding: 1.5rem 1.25rem 1.25rem;
    border-bottom: 1px solid var(--border, #e2e8f0);
    margin-bottom: 1.25rem;
  }
  
  .page-title {
    font-size: 1.5rem;
    font-weight: 800;
    margin: 0 0 0.5rem;
    color: #0f172a;
    font-family: 'Outfit', sans-serif;
  }
  
  .student-row {
    display: flex;
    align-items: center;
    gap: 1rem;
  }
  
  .student-avatar {
    width: 42px;
    height: 42px;
    border-radius: 50%;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 1.1rem;
  }
  
  .student-sub {
    font-size: 0.85rem;
    color: #64748b;
  }
  
  .student-sub b {
    color: #0f172a;
    font-size: 0.95rem;
  }
</style>
