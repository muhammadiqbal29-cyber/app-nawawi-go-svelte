<script>
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/auth.svelte.js';
  import GuruSidebar from '$lib/components/layout/GuruSidebar.svelte';
  import Header from '$lib/components/layout/Header.svelte';
  import BottomNav from '$lib/components/layout/BottomNav.svelte';
  import JenjangModal from '$lib/components/guru/JenjangModal.svelte';
  import TahunModal from '$lib/components/guru/TahunModal.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

  let { children } = $props();

  let sidebarOpen = $state(true);
  let currentPath = $derived($page.url.pathname);
  // fiturGuru sekarang di-derive dari aksesMap sekolah yang sedang aktif
  let fiturGuru = $derived(
    guruState.activeSekolahId ? guruState.aksesMap[guruState.activeSekolahId] || [] : []
  );

  $effect(() => {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

    if (auth.role === 'admin') {
      // Admin has full access to all features based on the school's features

      // Fetch all schools for admin
      fetch(`${apiUrl}/api/sekolah`, { credentials: 'include' })
        .then((res) => { if (!res.ok) throw new Error('HTTP ' + res.status); return res.json(); })
        .then((data) => {
          if (data && data.data) {
            guruState.sekolahList = data.data;
            const fullMap = {};
            guruState.sekolahList.forEach((s) => {
              let f = [];
              try {
                if (s.fitur) f = JSON.parse(s.fitur);
              } catch {}
              // Admin sees everything configured for the school, plus all default guru features
              if (!f.includes('cetak_rapor')) f.push('cetak_rapor');
              fullMap[s.id] = f;
            });
            guruState.aksesMap = fullMap;

            if (!guruState.activeSekolahId) {
              let savedId = null;
              if (typeof window !== 'undefined') {
                savedId = localStorage.getItem('siakad_active_sekolah_id');
              }
              const matchingSekolah = savedId
                ? guruState.sekolahList.find((s) => s.id == savedId)
                : null;

              if (matchingSekolah) {
                guruState.activeSekolahId = matchingSekolah.id;
                guruState.activeSekolahName = matchingSekolah.nama_sekolah;
              } else if (guruState.sekolahList.length === 1) {
                guruState.activeSekolahId = guruState.sekolahList[0].id;
                guruState.activeSekolahName = guruState.sekolahList[0].nama_sekolah;
                if (typeof window !== 'undefined') {
                  localStorage.setItem('siakad_active_sekolah_id', guruState.activeSekolahId);
                  localStorage.setItem('siakad_active_sekolah_name', guruState.activeSekolahName);
                }
              } else if (guruState.sekolahList.length > 1) {
                guruState.showJenjangModal = true;
              }
            }
          }
        })
        .catch((err) => console.error('Gagal mengambil jenjang admin', err));
    } else {
      async function loadGuruData() {
        try {
          // 1. Fetch Profil Guru to get guruSchoolAkses
          let guruSchoolAkses = {};
          const profilRes = await fetch(`${apiUrl}/api/guru/me/profil`, { credentials: 'include' });
          if (profilRes.ok) {
            const data = await profilRes.json();
            if (data && data.nama_lengkap) {
              auth.loggedInUser = data.nama_lengkap;
            }
            if (data && data.akses_sekolah) {
              try {
                guruSchoolAkses = JSON.parse(data.akses_sekolah) || {};
              } catch {}
            }
          }

          // 2. Fetch Jenjang Sekolah
          const sekolahRes = await fetch(`${apiUrl}/api/guru/sekolah`, { credentials: 'include' });
          if (sekolahRes.ok) {
            const data = await sekolahRes.json();
            if (data && data.data) {
              guruState.sekolahList = data.data;
              const map = {};
              guruState.sekolahList.forEach((s) => {
                // Get features assigned to the guru for this school
                let guruFeatures = guruSchoolAkses[s.id] || [];
                
                // Get features enabled for the school itself
                let schoolFeatures = [];
                try {
                  if (s.fitur) schoolFeatures = JSON.parse(s.fitur);
                } catch {}

                // Intersect guru features with school features (school features act as a master toggle)
                let f = guruFeatures.filter(feat => schoolFeatures.includes(feat));

                // Ensure cetak_rapor is always available
                if (!f.includes('cetak_rapor')) f.push('cetak_rapor');
                map[s.id] = f;
              });
              guruState.aksesMap = map;

              if (!guruState.activeSekolahId) {
                let savedId = null;
                if (typeof window !== 'undefined') {
                  savedId = localStorage.getItem('siakad_active_sekolah_id');
                }
                const matchingSekolah = savedId
                  ? guruState.sekolahList.find((s) => s.id == savedId)
                  : null;

                if (matchingSekolah) {
                  guruState.activeSekolahId = matchingSekolah.id;
                  guruState.activeSekolahName = matchingSekolah.nama_sekolah;
                } else if (guruState.sekolahList.length === 1) {
                  guruState.activeSekolahId = guruState.sekolahList[0].id;
                  guruState.activeSekolahName = guruState.sekolahList[0].nama_sekolah;
                  if (typeof window !== 'undefined') {
                    localStorage.setItem('siakad_active_sekolah_id', guruState.activeSekolahId);
                    localStorage.setItem('siakad_active_sekolah_name', guruState.activeSekolahName);
                  }
                } else if (guruState.sekolahList.length > 1) {
                  guruState.showJenjangModal = true;
                }
              }
            }
          }
        } catch (err) {
          console.error('Gagal mengambil data guru/sekolah', err);
        }
      }
      loadGuruData();
    }

    // Fetch Tahun Ajaran
    fetch(`${apiUrl}/api/guru/tahun-ajaran`, { credentials: 'include' })
      .then((res) => { if (!res.ok) throw new Error('HTTP ' + res.status); return res.json(); })
      .then((data) => {
        if (data && data.length > 0) {
          guruState.tahunList = data;
          
          let savedTaId = typeof window !== 'undefined' ? localStorage.getItem('siakad_active_ta_id') : null;
          let savedSemId = typeof window !== 'undefined' ? localStorage.getItem('siakad_active_sem_id') : null;
          let savedTaName = typeof window !== 'undefined' ? localStorage.getItem('siakad_active_ta_name') : null;

          if (savedTaId && savedSemId && savedTaName) {
            guruState.activeTahunId = parseInt(savedTaId);
            guruState.activeSemesterId = parseInt(savedSemId);
            guruState.activeTahunName = savedTaName;
          } else {
            const aktif = data.find((t) => t.status === 'Active') || data[0];
            let semAktif = null;
            if (aktif && aktif.semesters) {
              semAktif = aktif.semesters.find((s) => s.status_aktif) || aktif.semesters[0];
            }
            if (aktif && semAktif) {
              guruState.activeTahunId = aktif.id;
              guruState.activeSemesterId = semAktif.id;
              guruState.activeTahunName = `${aktif.nama_tahun} - ${semAktif.tipe_semester}`;
              
              if (typeof window !== 'undefined') {
                localStorage.setItem('siakad_active_ta_id', aktif.id);
                localStorage.setItem('siakad_active_sem_id', semAktif.id);
                localStorage.setItem('siakad_active_ta_name', guruState.activeTahunName);
              }
            }
          }
        }
      })
      .catch((err) => console.error('Gagal mengambil tahun ajaran aktif', err));
  });

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
  }

  function onLogout() {
    fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/logout`, {
      method: 'POST',
      credentials: 'include'
    }).finally(() => {
      auth.loggedInUser = null;
      auth.role = '';
      goto('/login');
    });
  }

  let currentTitle = $derived(
    currentPath.startsWith('/guru/input-diniyah')
      ? 'Nilai Diniyah'
      : currentPath.startsWith('/guru/input-nilai')
        ? 'Input Nilai'
        : currentPath.startsWith('/guru/rapor')
          ? 'Rapor'
          : currentPath.startsWith('/guru/materi-tahfiz')
            ? 'Konfigurasi Materi'
            : 'Beranda Guru'
  );
</script>

<div class="layout-wrapper" class:sidebar-closed={!sidebarOpen}>
  <GuruSidebar isOpen={sidebarOpen} onClose={() => sidebarOpen = false} {currentPath} {onLogout} {fiturGuru} userRole={auth.role} />
  
  {#if sidebarOpen}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="mobile-overlay" onclick={() => sidebarOpen = false}></div>
  {/if}

  <div class="main-content">
    <Header
      title={currentTitle}
      activeTahun={guruState.activeTahunName}
      onGantiTahun={() => (guruState.showTahunModal = true)}
      userRole="guru"
      activeJenjang={guruState.activeSekolahName}
      onGantiJenjang={() => (guruState.showJenjangModal = true)}
      userName={auth.loggedInUser || 'Guru'}
      onToggleSidebar={toggleSidebar}
      {onLogout}
    />

    <main class="content-area">
      {@render children()}
    </main>
  </div>
</div>

<JenjangModal />
<TahunModal />
<BottomNav {currentPath} {fiturGuru} />

<style>
  .layout-wrapper {
    display: flex;
    min-height: 100vh;
    width: 100%;
    background-color: var(--color-bg-body, #f8fafc);
    transition: padding-left 0.3s ease;
    padding-left: var(--sidebar-width, 260px);
  }
  .layout-wrapper.sidebar-closed {
    padding-left: 80px;
  }
  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .content-area {
    padding: 2rem;
    flex: 1;
  }
  .mobile-overlay {
    display: none;
  }
  @media (max-width: 768px) {
    .layout-wrapper,
    .layout-wrapper.sidebar-closed {
      padding-left: 0 !important;
    }
    .content-area {
      padding: 1rem;
      padding-bottom: 100px; /* Space for BottomNav */
    }
    /* Sembunyikan GuruSidebar di mobile secara paksa untuk diganti BottomNav */
    :global(.sidebar) {
      display: none !important;
    }
    .mobile-overlay {
      display: none !important; /* Hamburger is gone, no overlay needed */
    }
  }
</style>
