<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  let profilData = $state(null);
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    try {
      const res = await fetch(`${apiUrl}/api/guru/me/profil`, { credentials: 'include' });
      if (res.ok) {
        profilData = await res.json();
      } else {
        error = 'Gagal memuat profil.';
      }
    } catch {
      error = 'Terjadi kesalahan jaringan.';
    } finally {
      loading = false;
    }
  });

  function getFiturNames(jsonStr) {
    if (!jsonStr) return [];
    try {
      const arr = JSON.parse(jsonStr);
      return Array.isArray(arr) ? arr : [];
    } catch {
      return [];
    }
  }

  const fiturLabels = {
    presensi_mandiri: 'Presensi Mandiri',
    input_tahfiz: 'Input Nilai Tahfiz',
    cetak_rapor: 'Cetak Rapor'
  };
</script>

<div class="page-container">
  <div class="header">
    <div class="title-group">
      <div class="icon-wrapper">
        <Icon name="user" size={28} />
      </div>
      <div>
        <h1>Profil Saya</h1>
        <p>Lihat informasi akun dan hak akses Anda.</p>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="card loading">
      <Icon name="loader" size={24} class="spin" />
      Memuat data...
    </div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if profilData}
    <div class="profil-grid">
      <div class="card detail-card">
        <h2>Informasi Pribadi</h2>
        <div class="info-list">
          <div class="info-item">
            <span class="label">Nama Lengkap</span>
            <span class="value">{profilData.nama_lengkap}</span>
          </div>
          <div class="info-item">
            <span class="label">NIK / NIP</span>
            <span class="value">{profilData.nik || '-'}</span>
          </div>
          <div class="info-item">
            <span class="label">Email</span>
            <span class="value">{profilData.user?.email || '-'}</span>
          </div>
          <div class="info-item">
            <span class="label">No HP</span>
            <span class="value">{profilData.no_hp || '-'}</span>
          </div>
        </div>
      </div>

      <div class="side-panel">
        <div class="card bg-indigo">
          <h2>Jadwal Kerja</h2>
          <div class="time-box">
            <div class="time">
              <span class="t-label">Masuk</span>
              <span class="t-val">{profilData.jam_masuk || '-'}</span>
            </div>
            <div class="divider"></div>
            <div class="time">
              <span class="t-label">Pulang</span>
              <span class="t-val">{profilData.jam_pulang || '-'}</span>
            </div>
          </div>
        </div>

        <div class="card features-card">
          <h2>Hak Akses Anda</h2>
          <ul class="fitur-list">
            {#each getFiturNames(profilData.akses_fitur) as fitur (fitur)}
              <li>
                <Icon name="check" size={16} class="text-success" />
                {fiturLabels[fitur] || fitur}
              </li>
            {:else}
              <li class="empty-fitur">Belum ada hak akses khusus yang diberikan.</li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1000px;
    margin: 0 auto;
  }
  .header {
    margin-bottom: 2rem;
  }
  .title-group {
    display: flex;
    align-items: center;
    gap: 1rem;
  }
  .icon-wrapper {
    background: #e0e7ff;
    color: #4f46e5;
    padding: 1rem;
    border-radius: 1rem;
    display: flex;
  }
  h1 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: #1e293b;
  }
  .title-group p {
    margin: 0.25rem 0 0;
    color: #64748b;
  }

  .profil-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
  @media (min-width: 768px) {
    .profil-grid {
      grid-template-columns: 1fr 350px;
    }
  }

  .card {
    background: white;
    border-radius: 1rem;
    border: 1px solid #e2e8f0;
    padding: 1.5rem;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
  }
  .card h2 {
    margin: 0 0 1.5rem;
    font-size: 1.15rem;
    color: #1e293b;
    border-bottom: 1px solid #f1f5f9;
    padding-bottom: 0.75rem;
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }
  .info-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .info-item .label {
    font-size: 0.85rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .info-item .value {
    font-size: 1.05rem;
    color: #0f172a;
    font-weight: 500;
  }

  .side-panel {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .bg-indigo {
    background: linear-gradient(135deg, #4f46e5, #3b82f6);
    color: white;
    border: none;
  }
  .bg-indigo h2 {
    color: white;
    border-color: rgba(255, 255, 255, 0.2);
  }
  .time-box {
    display: flex;
    align-items: center;
    justify-content: space-around;
    padding: 0.5rem 0;
  }
  .time {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }
  .t-label {
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    opacity: 0.8;
  }
  .t-val {
    font-size: 1.75rem;
    font-weight: 700;
  }
  .divider {
    width: 2px;
    height: 40px;
    background: rgba(255, 255, 255, 0.2);
  }

  .fitur-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .fitur-list li {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-size: 0.95rem;
    color: #334155;
    background: #f8fafc;
    padding: 0.75rem 1rem;
    border-radius: 0.5rem;
    border: 1px solid #f1f5f9;
  }
  .fitur-list li :global(.text-success) {
    color: #10b981;
  }
  .empty-fitur {
    color: #64748b;
    font-size: 0.9rem;
    font-style: italic;
    background: none !important;
    border: none !important;
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    padding: 3rem;
    color: #64748b;
  }
  .spin {
    animation: spin 1s linear infinite;
  }
  @keyframes spin {
    100% {
      transform: rotate(360deg);
    }
  }
  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    padding: 1rem;
    border-radius: 0.5rem;
    border: 1px solid #fecaca;
  }
</style>
