<script>
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apps = [
    {
      id: 'institusi',
      title: 'Data Institusi',
      description: 'Kelola unit sekolah dan kalender akademik',
      icon: 'database',
      color: '#3b82f6', // blue
      path: '/institusi'
    },
    {
      id: 'akademik',
      title: 'Modul Akademik',
      description: 'Manajemen kelas, siswa, dan rombel (KBM)',
      icon: 'bookOpen',
      color: '#8b5cf6', // purple
      path: '/akademik'
    },
    {
      id: 'sdm',
      title: 'Manajemen SDM',
      description: 'Direktori guru dan pegawai yayasan',
      icon: 'users',
      color: '#10b981', // emerald
      path: '/sdm'
    },
    {
      id: 'presensi',
      title: 'Presensi & Kehadiran',
      description: 'Pantau log kehadiran dan ketidakhadiran',
      icon: 'clock',
      color: '#f59e0b', // amber
      path: '/presensi'
    }
  ];
</script>

<div class="page-container">
  <div class="header">
    <h1 class="page-title">Beranda Aplikasi</h1>
    <p class="page-subtitle">Pilih modul yang ingin Anda kerjakan hari ini.</p>
  </div>

  <div class="app-grid">
    {#each apps as app (app.id)}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="app-card" onclick={() => goto(app.path)}>
        <div class="icon-circle" style="background-color: {app.color}15; color: {app.color};">
          <Icon name={app.icon} size={32} />
        </div>
        <h3>{app.title}</h3>
        <p>{app.description}</p>
      </div>
    {/each}
  </div>
</div>

<style>
  .page-container {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }
  .header {
    text-align: center;
    margin-bottom: 3rem;
  }
  .page-title {
    font-size: 2rem;
    font-weight: 800;
    color: #0f172a;
    margin: 0 0 0.5rem 0;
  }
  .page-subtitle {
    color: #64748b;
    font-size: 1.1rem;
    margin: 0;
  }

  .app-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
  }

  .app-card {
    background: white;
    border-radius: 1rem;
    padding: 2rem;
    text-align: center;
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.05);
    border: 1px solid #e2e8f0;
    cursor: pointer;
    transition: all 0.3s ease;
  }
  .app-card:hover {
    transform: translateY(-5px);
    box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1);
    border-color: #cbd5e1;
  }

  .icon-circle {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0 auto 1.5rem auto;
  }

  .app-card h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.25rem;
    color: #1e293b;
  }
  .app-card p {
    margin: 0;
    color: #64748b;
    font-size: 0.95rem;
    line-height: 1.5;
  }
</style>
