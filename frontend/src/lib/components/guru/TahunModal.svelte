<script>
  import { guruState } from '$lib/guruState.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';

  function selectTahunSemester(ta, sem) {
    guruState.activeTahunId = ta.id;
    guruState.activeSemesterId = sem.id;
    guruState.activeTahunName = `${ta.nama_tahun} - ${sem.tipe_semester}`;
    guruState.showTahunModal = false;
    
    // Save to local storage to remember choice
    if (typeof window !== 'undefined') {
      localStorage.setItem('siakad_active_ta_id', ta.id);
      localStorage.setItem('siakad_active_sem_id', sem.id);
      localStorage.setItem('siakad_active_ta_name', guruState.activeTahunName);
    }
  }
</script>

{#if guruState.showTahunModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-backdrop" onclick={(e) => { if (e.target === e.currentTarget) guruState.showTahunModal = false; }}>
    <div class="modal-card">
      <div class="modal-header">
        <div class="icon-circle">
          <Icon name="calendar" size={24} />
        </div>
        <h2>Pilih Tahun Ajaran</h2>
        <p>Silakan pilih tahun ajaran dan semester yang ingin Anda kelola saat ini.</p>
      </div>

      <div class="modal-body">
        {#if guruState.tahunList.length === 0}
          <div class="empty-state">
            <p>Memuat tahun ajaran...</p>
          </div>
        {:else}
          <div class="tahun-grid">
            {#each guruState.tahunList as ta (ta.id)}
              <div class="ta-group">
                <div class="ta-header">
                  <span class="ta-name">{ta.nama_tahun}</span>
                  {#if ta.status === 'Active'}
                    <span class="badge-active">Aktif</span>
                  {/if}
                </div>
                
                <div class="sem-list">
                  {#each ta.semesters || [] as sem (sem.id)}
                    <button
                      class="sem-card {guruState.activeTahunId === ta.id && guruState.activeSemesterId === sem.id ? 'active' : ''}"
                      onclick={() => selectTahunSemester(ta, sem)}
                    >
                      <div class="sem-info">
                        <span class="sem-name">Semester {sem.tipe_semester}</span>
                      </div>
                      {#if guruState.activeTahunId === ta.id && guruState.activeSemesterId === sem.id}
                        <div class="check-icon">
                          <Icon name="check" size={16} />
                        </div>
                      {/if}
                    </button>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <div class="modal-footer">
        <button class="btn-tutup" onclick={() => (guruState.showTahunModal = false)}>
          Batal
        </button>
      </div>
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
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow:
      0 20px 25px -5px rgba(0, 0, 0, 0.1),
      0 10px 10px -5px rgba(0, 0, 0, 0.04);
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
    flex-shrink: 0;
  }
  .icon-circle {
    width: 56px;
    height: 56px;
    background: #ecfdf5;
    color: #10b981;
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
    overflow-y: auto;
  }
  .tahun-grid {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }
  .ta-group {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .ta-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .ta-name {
    font-weight: 700;
    font-size: 1.05rem;
    color: #1e293b;
  }
  .badge-active {
    background-color: #dcfce7;
    color: #15803d;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 700;
  }
  .sem-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .sem-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.85rem 1.25rem;
    background: white;
    border: 1px solid #cbd5e1;
    border-radius: 0.75rem;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s;
  }
  .sem-card:hover {
    border-color: #94a3b8;
    background: #f1f5f9;
  }
  .sem-card.active {
    border-color: #10b981;
    background: #ecfdf5;
    box-shadow: 0 4px 6px -1px rgba(16, 185, 129, 0.1);
  }
  .sem-name {
    font-weight: 600;
    color: #334155;
  }
  .sem-card.active .sem-name {
    color: #047857;
  }
  .check-icon {
    color: #10b981;
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
    flex-shrink: 0;
  }
  .btn-tutup {
    padding: 0.5rem 1rem;
    background: transparent;
    border: none;
    color: #64748b;
    font-weight: 600;
    cursor: pointer;
    border-radius: 0.5rem;
  }
  .btn-tutup:hover {
    background: #f1f5f9;
    color: #334155;
  }
</style>
