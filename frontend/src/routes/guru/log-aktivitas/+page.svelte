<script>
  import { onMount, untrack } from 'svelte';
  import { guruState } from '$lib/guruState.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let logs = $state([]);
  let loading = $state(false);
  let errorMsg = $state('');
  let page = $state(1);
  let limit = $state(20);
  let hasMore = $state(true);

  let activeSekolahId = $derived(guruState.activeSekolahId);

  async function fetchLogs(reset = false) {
    if (!activeSekolahId) return;
    if (reset) {
      page = 1;
      logs = [];
      hasMore = true;
    }

    if (!hasMore) return;

    loading = true;
    errorMsg = '';

    try {
      const res = await fetch(`${apiUrl}/api/guru/log-aktivitas?sekolah_id=${activeSekolahId}&page=${page}&limit=${limit}`, {
        credentials: 'include'
      });
      if (res.ok) {
        const body = await res.json();
        const newLogs = body.data || [];
        if (newLogs.length < limit) {
          hasMore = false;
        }
        logs = [...logs, ...newLogs];
      } else {
        const err = await res.json();
        errorMsg = err.message || 'Gagal mengambil log aktivitas';
      }
    } catch (e) {
      errorMsg = 'Kesalahan jaringan: ' + e.message;
    } finally {
      loading = false;
    }
  }

  function loadMore() {
    page += 1;
    fetchLogs();
  }

  function formatDate(isoString) {
    const d = new Date(isoString);
    return new Intl.DateTimeFormat('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(d);
  }

  function getIconForAksi(aksi) {
    const lower = aksi.toLowerCase();
    if (lower.includes('hapus') || lower.includes('delete')) return 'trash';
    if (lower.includes('edit') || lower.includes('perbaru') || lower.includes('autosave')) return 'edit';
    if (lower.includes('tambah') || lower.includes('input') || lower.includes('create')) return 'plus';
    return 'info';
  }

  $effect(() => {
    if (activeSekolahId) {
      untrack(() => {
        fetchLogs(true);
      });
    }
  });
</script>

<div class="page-container">
  <div class="header">
    <div class="title-group">
      <h1>Log Aktivitas Guru</h1>
      <p>Rekam jejak aktivitas Anda di sistem.</p>
    </div>
  </div>

  {#if errorMsg}
    <div class="alert alert-error">
      <Icon name="alertTriangle" size={20} />
      <span>{errorMsg}</span>
    </div>
  {/if}

  <div class="log-container">
    {#if logs.length === 0 && !loading}
      <div class="empty-state">
        <Icon name="clock" size={48} />
        <p>Belum ada aktivitas yang terekam.</p>
      </div>
    {/if}

    <div class="timeline">
      {#each logs as log (log.id)}
        <div class="timeline-item">
          <div class="timeline-icon">
            <Icon name={getIconForAksi(log.aksi)} size={16} />
          </div>
          <div class="timeline-content card">
            <div class="log-header">
              <div class="header-left">
                <span class="log-module badge">{log.modul}</span>
                {#if log.guru?.user?.nama_lengkap}
                  <span class="guru-name">
                    <Icon name="user" size={12} />
                    {log.guru.user.nama_lengkap}
                  </span>
                {/if}
              </div>
              <span class="log-date">{formatDate(log.updated_at)}</span>
            </div>
            <h3 class="log-action">{log.aksi}</h3>
            <p class="log-desc">{log.deskripsi}</p>
          </div>
        </div>
      {/each}
    </div>

    {#if loading}
      <div class="loading">
        <div class="spinner"></div>
        <span>Memuat log...</span>
      </div>
    {/if}

    {#if hasMore && !loading && logs.length > 0}
      <button class="btn-load-more" onclick={loadMore}>Muat Lebih Banyak</button>
    {/if}
  </div>
</div>

<style>
  .page-container {
    max-width: 800px;
    margin: 0 auto;
    padding-bottom: 2rem;
  }
  .header {
    margin-bottom: 2rem;
  }
  .title-group h1 {
    margin: 0;
    font-size: 1.5rem;
    color: #1e293b;
  }
  .title-group p {
    margin: 0.2rem 0 0;
    color: #64748b;
  }
  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    padding: 1rem;
    border-radius: 0.5rem;
    border: 1px solid #fecaca;
    margin-bottom: 1.5rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .empty-state {
    text-align: center;
    padding: 4rem 1rem;
    color: #94a3b8;
  }
  .empty-state :global(svg) {
    margin-bottom: 1rem;
    opacity: 0.5;
  }
  
  .timeline {
    position: relative;
    padding-left: 24px;
  }
  
  .timeline::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: 7px;
    width: 2px;
    background: #e2e8f0;
  }

  .timeline-item {
    position: relative;
    margin-bottom: 1.5rem;
  }

  .timeline-icon {
    position: absolute;
    left: -24px;
    top: 0;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #4f46e5;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 4px solid #f8fafc;
  }

  .card {
    background: white;
    border-radius: 0.75rem;
    padding: 1rem 1.25rem;
    border: 1px solid #e2e8f0;
    box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  }

  .log-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }
  .header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .guru-name {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.8rem;
    font-weight: 500;
    color: #475569;
    background: #f1f5f9;
    padding: 0.2rem 0.5rem;
    border-radius: 0.25rem;
  }
  .badge {
    background: #eef2ff;
    color: #4f46e5;
    padding: 0.25rem 0.5rem;
    border-radius: 0.375rem;
    font-size: 0.75rem;
    font-weight: 600;
  }
  .log-date {
    font-size: 0.8rem;
    color: #64748b;
  }
  .log-action {
    margin: 0 0 0.5rem 0;
    font-size: 1.05rem;
    color: #1e293b;
  }
  .log-desc {
    margin: 0;
    font-size: 0.9rem;
    color: #475569;
    line-height: 1.4;
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    color: #64748b;
    padding: 1rem 0;
  }
  .spinner {
    width: 1rem;
    height: 1rem;
    border: 2px solid #e2e8f0;
    border-top-color: #4f46e5;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .btn-load-more {
    display: block;
    width: 100%;
    padding: 0.75rem;
    background: white;
    border: 1px dashed #cbd5e1;
    color: #4f46e5;
    font-weight: 600;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.2s;
    margin-top: 1rem;
  }
  .btn-load-more:hover {
    background: #f8fafc;
    border-color: #94a3b8;
  }
</style>
