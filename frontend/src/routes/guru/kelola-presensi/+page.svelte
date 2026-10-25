<script>
  import { guruState } from '$lib/guruState.svelte.js';

  const api = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  let teachers = $state([]);
  let rows = $state([]);
  let teacherId = $state('');
  let month = $state(new Date().getMonth() + 1);
  let year = $state(new Date().getFullYear());
  let busy = $state(false);
  let message = $state('');
  let error = $state('');
  let config = $state({check_in_from:'07:00',check_in_to:'07:30',check_out_from:'15:00',check_out_to:'16:00',skip_sundays:true,overwrite_existing:false});
  const schoolId = $derived(guruState.activeSekolahId);

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
  async function generate() {
    busy = true; error = ''; message = '';
    try {
      const res = await fetch(`${api}/api/guru/presensi/kelola/generate-bulan?sekolah_id=${encodeURIComponent(schoolId)}`, {
        method:'POST', credentials:'include', headers:{'Content-Type':'application/json'},
        body:JSON.stringify({teacher_id:Number(teacherId),bulan:Number(month),tahun:Number(year),...config})
      });
      const data = await res.json().catch(()=>({}));
      if (!res.ok) throw new Error(data.message || 'Gagal membuat absensi bulanan');
      const result = data.data || {};
      message = `Selesai: ${result.created || 0} dibuat, ${result.updated || 0} diperbarui, ${result.skipped || 0} dilewati.`;
      await load();
    } catch(e) { error = e.message; } finally { busy = false; }
  }

  async function exportXlsx() {
    if (!schoolId) return;
    try {
      const qs = new URLSearchParams({ sekolah_id: String(schoolId), month: String(month), year: String(year) });
      if (teacherId) qs.set('teacher_id', String(teacherId));
      const res = await fetch(`${api}/api/guru/presensi/kelola/export?${qs.toString()}`, { credentials: 'include' });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.message || 'Gagal mengunduh file absensi');
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      const suffix = teacherId ? `_${teacherId}` : '';
      link.download = `absensi_guru_${year}_${String(month).padStart(2, '0')}${suffix}.xlsx`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      message = 'File Excel absensi berhasil diunduh.';
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
      <h2>Generate Absensi Bulanan</h2>
      <div class="grid">
        <label>Guru<select bind:value={teacherId}><option value="">Pilih guru</option>{#each teachers as teacher}<option value={teacher.id}>{teacher.nama_lengkap}</option>{/each}</select></label>
        <label>Bulan<input type="number" min="1" max="12" bind:value={month} /></label>
        <label>Tahun<input type="number" min="2000" bind:value={year} /></label>
        <label>Masuk mulai<input type="time" bind:value={config.check_in_from} /></label>
        <label>Masuk sampai<input type="time" bind:value={config.check_in_to} /></label>
        <label>Pulang mulai<input type="time" bind:value={config.check_out_from} /></label>
        <label>Pulang sampai<input type="time" bind:value={config.check_out_to} /></label>
      </div>
      <label class="check"><input type="checkbox" bind:checked={config.skip_sundays}/> Lewati hari Minggu</label>
      <label class="check"><input type="checkbox" bind:checked={config.overwrite_existing}/> Timpa absensi yang sudah ada</label>
      <button disabled={busy || !teacherId} onclick={generate}>{busy ? 'Memproses…' : 'Generate Absensi Bulanan'}</button>
    </section>
    <section class="card">
      <div class="row">
        <h2>Riwayat Absensi Sekolah</h2>
        <div class="row-actions">
          <button class="secondary" onclick={exportXlsx}>Download Excel</button>
          <button class="secondary" onclick={load}>Muat ulang</button>
        </div>
      </div>
      {#if rows.length === 0}<p>Belum ada data untuk bulan ini.</p>{:else}
        <div class="table-wrap"><table><thead><tr><th>Tanggal</th><th>Guru</th><th>Status</th><th>Masuk</th><th>Pulang</th></tr></thead><tbody>
          {#each rows as row}<tr><td>{row.tanggal}</td><td>{row.nama_lengkap}</td><td>{row.status}</td><td>{row.waktu_masuk || '—'}</td><td>{row.waktu_pulang || '—'}</td></tr>{/each}
        </tbody></table></div>
      {/if}
    </section>
  {/if}
</div>

<style>
  .page{max-width:1100px;margin:auto}h1{margin:0 0 .4rem}.card{background:white;border:1px solid #e2e8f0;border-radius:14px;padding:1.25rem;margin:1rem 0}.card h2{font-size:1.1rem;margin:0 0 1rem}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:1rem}label{display:grid;gap:.4rem;font-size:.9rem}input,select{padding:.65rem;border:1px solid #cbd5e1;border-radius:8px}.check{display:flex;margin-top:1rem;align-items:center}button{margin-top:1rem;background:#2563eb;color:#fff;border:0;border-radius:8px;padding:.7rem 1rem;cursor:pointer}button:disabled{opacity:.6}.secondary{margin:0;background:#475569}.row{display:flex;align-items:center;justify-content:space-between}.row-actions{display:flex;gap:.5rem;align-items:center;flex-wrap:wrap}.alert{padding:.8rem;border-radius:8px}.error{background:#fee2e2;color:#991b1b}.success{background:#dcfce7;color:#166534}.table-wrap{overflow:auto}table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:.7rem;border-bottom:1px solid #e2e8f0}@media(max-width:600px){.card{padding:1rem}.row{display:block}.row-actions{margin-top:.75rem}}
</style>
