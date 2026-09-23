<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';
  
  // Ambil API URL dari environment (VITE_API_URL) atau biarkan kosong jika relative
  const apiUrl = import.meta.env.VITE_API_URL || '';

  let guruList = $state([]);
  let loading = $state(false);

  let genBulanForm = $state({
    teacher_id: 0,
    bulan: parseInt(new Date().getMonth() + 1),
    tahun: new Date().getFullYear(),
    check_in_from: '06:30',
    check_in_to: '07:15',
    check_out_from: '15:30',
    check_out_to: '16:30',
    overwrite_existing: false,
    skip_sundays: true
  });

  async function fetchWithTimeout(url, options = {}, timeout = 10000) {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);
    options.signal = controller.signal;
    const res = await fetch(url, options);
    clearTimeout(id);
    return res;
  }

  onMount(async () => {
    await loadGuru();
  });

  async function loadGuru() {
    try {
      const res = await fetchWithTimeout(`${apiUrl}/api/guru?limit=1000`, { credentials: 'include' });
      if (res.ok) {
        const data = await res.json();
        guruList = Array.isArray(data) ? data : (data.data || []);
        if (guruList.length > 0 && genBulanForm.teacher_id === 0) {
          genBulanForm.teacher_id = guruList[0].id;
        }
      }
    } catch (e) {
      console.error("Gagal memuat daftar guru", e);
    }
  }

  async function submitGenerateBulan(e) {
    e.preventDefault();
    
    if (genBulanForm.teacher_id == 0) {
      alert("Silakan pilih guru terlebih dahulu.");
      return;
    }
    
    const isConfirm = confirm(`Yakin ingin men-generate presensi 1 bulan?`);
    if (!isConfirm) return;

    loading = true;

    try {
      const res = await fetchWithTimeout(`${apiUrl}/api/presensi/generate-acak-bulan`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          teacher_id: parseInt(genBulanForm.teacher_id),
          bulan: parseInt(genBulanForm.bulan),
          tahun: parseInt(genBulanForm.tahun),
          check_in_from: genBulanForm.check_in_from,
          check_in_to: genBulanForm.check_in_to,
          check_out_from: genBulanForm.check_out_from,
          check_out_to: genBulanForm.check_out_to,
          overwrite_existing: genBulanForm.overwrite_existing,
          skip_sundays: genBulanForm.skip_sundays
        }),
        credentials: 'include'
      });
      
      const body = await res.json().catch(() => ({}));
      
      if (res.ok) {
        const meta = body.data || {};
        alert(`Generate absensi ${meta.teacher_name || "guru"} bulan ${genBulanForm.bulan}/${genBulanForm.tahun} selesai.\nDibuat ${meta.created || 0}, Diperbarui ${meta.updated || 0}, Dilewati ${meta.skipped || 0}.`);
        goto('/presensi');
      } else {
        alert(body.message || 'Gagal men-generate presensi bulanan.');
      }
    } catch (err) {
      alert('Terjadi kesalahan koneksi.');
    } finally {
      loading = false;
    }
  }
</script>

<div class="page-container">
  <div class="header-section">
    <button class="btn-secondary" onclick={() => goto('/presensi')} style="display:flex; align-items:center; gap: 8px;">
      <Icon name="arrow-left" size={16} /> Kembali
    </button>
    <div style="margin-left: 16px;">
      <h2 style="margin: 0; font-size: 1.5rem;">Generate 1 Bulan (Acak Presisi)</h2>
      <p class="page-subtitle">Pilih rentang waktu untuk menghasilkan data absensi otomatis bagi seorang guru.</p>
    </div>
  </div>

  <div class="card">
    <form class="form-grid" onsubmit={submitGenerateBulan}>
      
      <!-- Baris 1: Guru -->
      <div class="form-group full-width">
        <label>Guru Target</label>
        <select class="form-input" bind:value={genBulanForm.teacher_id} required>
          {#each guruList as g}
            <option value={g.id}>{g.nama_lengkap}</option>
          {/each}
        </select>
        <small>Pilih guru yang ingin dibuatkan absensi fiktifnya.</small>
      </div>
      
      <!-- Baris 2: Periode -->
      <div class="form-group">
        <label>Bulan (1-12)</label>
        <input type="number" class="form-input" bind:value={genBulanForm.bulan} min="1" max="12" required/>
      </div>
      <div class="form-group">
        <label>Tahun (YYYY)</label>
        <input type="number" class="form-input" bind:value={genBulanForm.tahun} required/>
      </div>

      <!-- Baris 3: Jam Masuk -->
      <div class="form-group">
        <label>Check In From (Mulai)</label>
        <input type="time" class="form-input" bind:value={genBulanForm.check_in_from} required/>
        <small>Waktu paling awal untuk absensi masuk.</small>
      </div>
      <div class="form-group">
        <label>Check In To (Hingga)</label>
        <input type="time" class="form-input" bind:value={genBulanForm.check_in_to} required/>
        <small>Waktu paling akhir untuk absensi masuk.</small>
      </div>

      <!-- Baris 4: Jam Pulang -->
      <div class="form-group">
        <label>Check Out From (Mulai)</label>
        <input type="time" class="form-input" bind:value={genBulanForm.check_out_from} required/>
        <small>Waktu paling awal untuk absensi pulang.</small>
      </div>
      <div class="form-group">
        <label>Check Out To (Hingga)</label>
        <input type="time" class="form-input" bind:value={genBulanForm.check_out_to} required/>
        <small>Waktu paling akhir untuk absensi pulang.</small>
      </div>

      <!-- Baris 5: Opsi Lanjutan -->
      <div class="form-group full-width checkbox-group">
        <div class="checkbox-item">
          <input type="checkbox" id="skipSun" bind:checked={genBulanForm.skip_sundays} />
          <label for="skipSun">Lewati hari Minggu (Skip Sundays)</label>
        </div>
        <div class="checkbox-item">
          <input type="checkbox" id="overwriteEx" bind:checked={genBulanForm.overwrite_existing} />
          <label for="overwriteEx">Timpa data yang sudah ada (Overwrite Existing)</label>
        </div>
      </div>

      <!-- Baris 6: Aksi -->
      <div class="form-group full-width" style="text-align: right; margin-top: 1rem;">
        <button type="submit" class="btn-primary" disabled={loading} style="padding: 12px 24px; font-size: 1rem;">
          <Icon name="zap" size={16} /> {loading ? 'Menyimpan...' : 'Generate Presensi'}
        </button>
      </div>

    </form>
  </div>
</div>

<style>
  .page-container {
    padding: 1rem;
    max-width: 900px;
    margin: 0 auto;
  }
  .header-section {
    display: flex;
    align-items: center;
    margin-bottom: 2rem;
  }
  .page-subtitle {
    color: #6b7280;
    margin: 4px 0 0 0;
    font-size: 0.95rem;
  }
  .card {
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    padding: 2rem;
  }
  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .full-width {
    grid-column: 1 / -1;
  }
  .form-group label {
    font-weight: 500;
    font-size: 0.9rem;
    color: #374151;
  }
  .form-input {
    padding: 0.6rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.95rem;
  }
  .form-input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }
  small {
    color: #6b7280;
    font-size: 0.8rem;
  }
  .checkbox-group {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 0.5rem;
  }
  .checkbox-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .checkbox-item input {
    width: 16px;
    height: 16px;
    cursor: pointer;
  }
  .checkbox-item label {
    cursor: pointer;
    font-weight: normal;
  }
  
  .btn-primary {
    background: #2563eb;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 6px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .btn-primary:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  .btn-secondary {
    background: #e5e7eb;
    color: #374151;
    border: none;
    padding: 8px 16px;
    border-radius: 6px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .btn-secondary:hover {
    background: #d1d5db;
  }
</style>
