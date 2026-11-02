<script>
  import { goto } from '$app/navigation';
  import { guruState } from '$lib/guruState.svelte.js';

  const api = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  const schoolId = $derived(guruState.activeSekolahId);

  let teachers = $state([]);
  let teacherId = $state('');
  let teacherQuery = $state('');
  let showTeacherMenu = $state(false);
  let busy = $state(false);
  let error = $state('');
  let message = $state('');
  let config = $state({
    check_in_from: '07:00',
    check_in_to: '07:30',
    check_out_from: '15:00',
    check_out_to: '16:00',
    skip_sundays: true,
    overwrite_existing: false,
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear()
  });

  let filteredTeachers = $derived.by(() => {
    const q = teacherQuery.trim().toLowerCase();
    if (!q) return teachers;
    return teachers.filter((teacher) => teacher.nama_lengkap.toLowerCase().includes(q));
  });

  async function loadTeachers() {
    if (!schoolId) return;
    try {
      const res = await fetch(`${api}/api/guru/presensi/kelola/guru?sekolah_id=${encodeURIComponent(schoolId)}`, {
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal memuat daftar guru');
      teachers = await res.json();
      if (!teacherId && teachers.length > 0) {
        teacherId = String(teachers[0].id);
        teacherQuery = teachers[0].nama_lengkap;
      }
    } catch (e) {
      error = e.message;
    }
  }

  $effect(() => {
    if (schoolId) loadTeachers();
  });

  $effect(() => {
    if (teacherId) {
      const selectedTeacher = teachers.find((teacher) => String(teacher.id) === String(teacherId));
      if (selectedTeacher && teacherQuery !== selectedTeacher.nama_lengkap) {
        teacherQuery = selectedTeacher.nama_lengkap;
      }
    }
  });

  async function generate() {
    if (!schoolId || !teacherId) {
      error = 'Pilih guru terlebih dahulu.';
      return;
    }

    busy = true;
    error = '';
    message = '';

    try {
      const res = await fetch(`${api}/api/guru/presensi/kelola/generate-bulan?sekolah_id=${encodeURIComponent(schoolId)}`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          teacher_id: Number(teacherId),
          bulan: Number(config.month),
          tahun: Number(config.year),
          check_in_from: config.check_in_from,
          check_in_to: config.check_in_to,
          check_out_from: config.check_out_from,
          check_out_to: config.check_out_to,
          skip_sundays: config.skip_sundays,
          overwrite_existing: config.overwrite_existing
        })
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || 'Gagal membuat absensi bulanan');

      const result = data.data || {};
      message = `Selesai: ${result.created || 0} dibuat, ${result.updated || 0} diperbarui, ${result.skipped || 0} dilewati.`;
      setTimeout(() => goto('/guru/kelola-presensi'), 400);
    } catch (e) {
      error = e.message;
    } finally {
      busy = false;
    }
  }
</script>

<svelte:head>
  <title>Generate Absensi</title>
</svelte:head>

<div class="page">
  <div class="topbar">
    <button type="button" class="secondary" onclick={() => goto('/guru/kelola-presensi')}>Kembali</button>
    <h1>Generate Absensi Bulanan</h1>
  </div>

  {#if !schoolId}
    <div class="card">Pilih sekolah aktif terlebih dahulu.</div>
  {:else}
    <section class="card">
      {#if error}<div class="alert error">{error}</div>{/if}
      {#if message}<div class="alert success">{message}</div>{/if}

      <div class="grid">
        <div class="teacher-picker">
          <label>Guru</label>
          <input
            type="text"
            value={teacherQuery}
            placeholder="Cari guru..."
            oninput={(event) => {
              teacherQuery = event.currentTarget.value;
              showTeacherMenu = true;
            }}
            onfocus={() => (showTeacherMenu = true)}
          />
          {#if showTeacherMenu}
            <div class="teacher-menu" role="listbox" aria-label="Daftar guru generate">
              <button type="button" class:active={!teacherId} onclick={() => { teacherId = ''; teacherQuery = ''; showTeacherMenu = false; }}>
                Pilih guru
              </button>
              {#if filteredTeachers.length > 0}
                {#each filteredTeachers as teacher}
                  <button
                    type="button"
                    class:active={String(teacher.id) === String(teacherId)}
                    onclick={() => {
                      teacherId = String(teacher.id);
                      teacherQuery = teacher.nama_lengkap;
                      showTeacherMenu = false;
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

        <label>
          Bulan
          <input type="number" min="1" max="12" bind:value={config.month} />
        </label>

        <label>
          Tahun
          <input type="number" min="2000" bind:value={config.year} />
        </label>

        <label>
          Masuk mulai
          <input type="time" bind:value={config.check_in_from} />
        </label>

        <label>
          Masuk sampai
          <input type="time" bind:value={config.check_in_to} />
        </label>

        <label>
          Pulang mulai
          <input type="time" bind:value={config.check_out_from} />
        </label>

        <label>
          Pulang sampai
          <input type="time" bind:value={config.check_out_to} />
        </label>
      </div>

      <label class="check"><input type="checkbox" bind:checked={config.skip_sundays} /> Lewati hari Minggu</label>
      <label class="check"><input type="checkbox" bind:checked={config.overwrite_existing} /> Timpa absensi yang sudah ada</label>

      <button type="button" class="primary" disabled={busy || !teacherId} onclick={generate}>
        {busy ? 'Memproses…' : 'Generate Absensi'}
      </button>
    </section>
  {/if}
</div>

<style>
  .page {
    max-width: 1100px;
    margin: 0 auto;
    padding: 1rem;
  }

  .topbar {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  h1 {
    margin: 0;
    font-size: 1.6rem;
  }

  .card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 14px;
    padding: 1.25rem;
    margin: 1rem 0;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
  }

  label {
    display: grid;
    gap: 0.45rem;
    font-size: 0.9rem;
  }

  .teacher-picker {
    position: relative;
    display: grid;
    gap: 0.45rem;
    font-size: 0.9rem;
  }

  .teacher-picker input {
    width: 100%;
    padding: 0.7rem;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    font: inherit;
  }

  .teacher-menu {
    position: absolute;
    top: calc(100% + 0.35rem);
    left: 0;
    right: 0;
    z-index: 10;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 10px;
    box-shadow: 0 10px 26px rgba(15, 23, 42, 0.12);
    max-height: 220px;
    overflow: auto;
  }

  .teacher-menu button {
    display: block;
    width: 100%;
    text-align: left;
    padding: 0.7rem 0.8rem;
    background: transparent;
    color: #1e293b;
    border: 0;
    border-bottom: 1px solid #f1f5f9;
    cursor: pointer;
    margin: 0;
    border-radius: 0;
  }

  .teacher-menu button:last-child {
    border-bottom: 0;
  }

  .teacher-menu button:hover,
  .teacher-menu button.active {
    background: #eff6ff;
    color: #0f172a;
  }

  .teacher-empty {
    padding: 0.8rem;
    color: #64748b;
    font-size: 0.85rem;
  }

  input,
  select {
    padding: 0.7rem;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    font: inherit;
  }

  .check {
    display: flex;
    align-items: center;
    margin-top: 1rem;
  }

  .primary,
  .secondary {
    margin-top: 1.2rem;
    border: 0;
    border-radius: 8px;
    padding: 0.8rem 1rem;
    cursor: pointer;
  }

  .primary {
    background: #2563eb;
    color: white;
  }

  .secondary {
    background: #475569;
    color: white;
  }

  button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .alert {
    padding: 0.8rem;
    border-radius: 8px;
    margin-bottom: 1rem;
  }

  .alert.error {
    background: #fee2e2;
    color: #991b1b;
  }

  .alert.success {
    background: #dcfce7;
    color: #166534;
  }
</style>
