<script>
  import { goto } from '$app/navigation';
  import { guruState } from '$lib/guruState.svelte.js';

  const api = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  let teachers = $state([]);
  let rows = $state([]);
  let teacherId = $state('');
  let exportTeacherId = $state('');
  let month = $state(new Date().getMonth() + 1);
  let year = $state(new Date().getFullYear());
  let busy = $state(false);
  let message = $state('');
  let error = $state('');
  let config = $state({check_in_from:'07:00',check_in_to:'07:30',check_out_from:'15:00',check_out_to:'16:00',skip_sundays:true,overwrite_existing:false});
  let teacherQuery = $state('');
  let exportTeacherQuery = $state('');
  let showTeacherMenu = $state(false);
  let showExportTeacherMenu = $state(false);
  const schoolId = $derived(guruState.activeSekolahId);

  function formatTanggal(value) {
    if (!value) return '—';
    const raw = new Date(value.includes('T') ? value : `${value}T00:00:00`);
    if (Number.isNaN(raw.getTime())) return String(value);
    return new Intl.DateTimeFormat('id-ID', {
      day: '2-digit',
      month: 'long',
      year: 'numeric'
    }).format(raw);
  }

  function getTeacherNameById(id) {
    return teachers.find((teacher) => String(teacher.id) === String(id))?.nama_lengkap || '';
  }

  let filteredTeachers = $derived.by(() => {
    const q = teacherQuery.trim().toLowerCase();
    if (!q) return teachers;
    return teachers.filter((teacher) => teacher.nama_lengkap.toLowerCase().includes(q));
  });

  let filteredExportTeachers = $derived.by(() => {
    const q = exportTeacherQuery.trim().toLowerCase();
    if (!q) return teachers;
    return teachers.filter((teacher) => teacher.nama_lengkap.toLowerCase().includes(q));
  });

  $effect(() => {
    if (teacherId) {
      const selectedName = getTeacherNameById(teacherId);
      if (selectedName && teacherQuery !== selectedName) {
        teacherQuery = selectedName;
      }
    }
    if (!exportTeacherId && teacherId) {
      exportTeacherId = teacherId;
      exportTeacherQuery = getTeacherNameById(teacherId);
    }
    if (exportTeacherId) {
      const selectedName = getTeacherNameById(exportTeacherId);
      if (selectedName && exportTeacherQuery !== selectedName) {
        exportTeacherQuery = selectedName;
      }
    }
  });

  $effect(() => { if (schoolId) load(); });
  async function load() {
    if (!schoolId) return;
    error = '';
    try {
      const qs = `sekolah_id=${encodeURIComponent(schoolId)}`;
      const [tRes, rRes] = await Promise.all([
        fetch(`${api}/api/guru/presensi/kelola/guru?${qs}`, {credentials:'include'}),
        fetch(`${api}/api/guru/presensi/kelola/riwayat?${qs}&month=${month}&year=${year}`, {credentials:'include'})
      ]);
      if (!tRes.ok || !rRes.ok) throw new Error('Tidak dapat memuat data pengelolaan absensi. Periksa hak akses sekolah ini.');
      teachers = await tRes.json(); rows = await rRes.json();
    } catch(e) { error = e.message; }
  }
  async function exportXlsx() {
    if (!schoolId) return;
    try {
      const qs = new URLSearchParams({ sekolah_id: String(schoolId), month: String(month), year: String(year) });
      const selectedTeacherId = exportTeacherId || teacherId;
      if (selectedTeacherId) qs.set('teacher_id', String(selectedTeacherId));
      const res = await fetch(`${api}/api/guru/presensi/kelola/export?${qs.toString()}`, { credentials: 'include' });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.message || 'Gagal mengunduh file absensi');
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      const suffix = selectedTeacherId ? `_${selectedTeacherId}` : '_semua_guru';
      link.download = `absensi_guru_${year}_${String(month).padStart(2, '0')}${suffix}.xlsx`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      const targetName = selectedTeacherId ? getTeacherNameById(selectedTeacherId) : 'Semua Guru';
      message = `File Excel absensi ${targetName} berhasil diunduh.`;
    } catch (e) {
      error = e.message;
    }
  }

</script>

<svelte:head><title>Kelola Absensi</title></svelte:head>
<div class="page">
  <h1>Kelola Absensi</h1>
  <p>Sekolah aktif: <strong>{guruState.activeSekolahName || 'Pilih sekolah'}</strong></p>
  {#if error}<div class="alert error">{error}</div>{/if}
  {#if message}<div class="alert success">{message}</div>{/if}
  {#if !schoolId}<div class="card">Pilih sekolah aktif terlebih dahulu.</div>{:else}
    <section class="card">
      <div class="row">
        <h2>Generate Absensi</h2>
        <button class="secondary" onclick={() => goto('/guru/kelola-presensi/generate')}>Buat absensi baru</button>
      </div>
      <p class="muted">Untuk membuat absensi bulanan, buka halaman terpisah agar prosesnya tidak tercampur dengan riwayat dan export.</p>
    </section>
    <section class="card">
      <div class="row">
        <h2>Riwayat Absensi Sekolah</h2>
        <div class="row-actions">
          <div class="export-choice">
            <span>Unduh untuk</span>
            <div class="export-picker">
              <input
                type="text"
                value={exportTeacherQuery}
                placeholder="Cari guru..."
                oninput={(event) => {
                  exportTeacherQuery = event.currentTarget.value;
                  showExportTeacherMenu = true;
                }}
                onfocus={() => (showExportTeacherMenu = true)}
              />
              {#if showExportTeacherMenu}
                <div class="teacher-menu" role="listbox" aria-label="Daftar guru export">
                  <button type="button" class:active={!exportTeacherId} onclick={() => { exportTeacherId = ''; exportTeacherQuery = ''; showExportTeacherMenu = false; }}>
                    Semua guru
                  </button>
                  {#if filteredExportTeachers.length > 0}
                    {#each filteredExportTeachers as teacher}
                      <button
                        type="button"
                        class:active={String(teacher.id) === String(exportTeacherId)}
                        onclick={() => {
                          exportTeacherId = String(teacher.id);
                          exportTeacherQuery = teacher.nama_lengkap;
                          showExportTeacherMenu = false;
                        }}
                      >
                        {teacher.nama_lengkap}
                      </button>
                    {/each}
                  {:else}
                    <div class="teacher-empty">Tidak ada guru yang cocok</div>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
          <button class="secondary" onclick={exportXlsx}>Download Excel</button>
          <button class="secondary" onclick={load}>Muat ulang</button>
        </div>
      </div>
      {#if rows.length === 0}<p>Belum ada data untuk bulan ini.</p>{:else}
        <div class="table-shell">
          <table><thead><tr><th>Tanggal</th><th>Guru</th><th>Status</th><th>Masuk</th><th>Pulang</th></tr></thead><tbody>
            {#each rows as row}<tr><td>{formatTanggal(row.tanggal)}</td><td>{row.nama_lengkap}</td><td>{row.status}</td><td>{row.waktu_masuk || '—'}</td><td>{row.waktu_pulang || '—'}</td></tr>{/each}
          </tbody></table>
        </div>
      {/if}
    </section>
  {/if}
</div>

<style>
  .page{max-width:1100px;margin:auto}h1{margin:0 0 .4rem}.card{background:white;border:1px solid #e2e8f0;border-radius:14px;padding:1.25rem;margin:1rem 0}.card h2{font-size:1.1rem;margin:0 0 1rem}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:1rem}label{display:grid;gap:.4rem;font-size:.9rem}input,select{padding:.65rem;border:1px solid #cbd5e1;border-radius:8px}.teacher-picker{position:relative}.teacher-picker input,.export-picker input{width:100%}.teacher-menu{position:absolute;top:calc(100% + .35rem);left:0;right:0;z-index:10;background:#fff;border:1px solid #cbd5e1;border-radius:10px;box-shadow:0 10px 26px rgba(15,23,42,.12);max-height:220px;overflow:auto}.teacher-menu button{display:block;width:100%;text-align:left;padding:.7rem .8rem;background:transparent;color:#1e293b;border:0;border-bottom:1px solid #f1f5f9;cursor:pointer;margin:0;border-radius:0}.teacher-menu button:last-child{border-bottom:0}.teacher-menu button:hover,.teacher-menu button.active{background:#eff6ff;color:#0f172a}.teacher-empty{padding:.8rem;color:#64748b;font-size:.85rem}.export-choice{display:flex;align-items:center;gap:.5rem;min-width:290px}.export-choice span{font-size:.8rem;color:#475569}.export-picker{position:relative;min-width:200px}.check{display:flex;margin-top:1rem;align-items:center}.muted{margin-top:.5rem;color:#64748b;font-size:.9rem}.button,button{margin-top:1rem;background:#2563eb;color:#fff;border:0;border-radius:8px;padding:.7rem 1rem;cursor:pointer}button:disabled{opacity:.6}.secondary{margin:0;background:#475569}.row{display:flex;align-items:center;justify-content:space-between}.row-actions{display:flex;gap:.5rem;align-items:center;flex-wrap:wrap}.alert{padding:.8rem;border-radius:8px}.error{background:#fee2e2;color:#991b1b}.success{background:#dcfce7;color:#166534}.table-shell{width:100%;overflow-x:auto;overflow-y:visible}.table-shell table{min-width:760px;width:100%;border-collapse:collapse}th,td{text-align:left;padding:.7rem;border-bottom:1px solid #e2e8f0;white-space:nowrap}@media(max-width:600px){.card{padding:1rem}.row{display:block}.row-actions{margin-top:.75rem}.export-choice{width:100%;display:block}.export-picker{margin-top:.5rem}} 
</style>
