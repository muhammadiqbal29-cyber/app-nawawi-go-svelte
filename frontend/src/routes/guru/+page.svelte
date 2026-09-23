<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import BarChart from '$lib/components/charts/BarChart.svelte';
  import PresensiWidget from '$lib/components/guru/PresensiWidget.svelte';
  import { guruState } from '$lib/guruState.svelte.js';
  import '$lib/components/guru/AlquranMobileUI.css';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let selectedSekolah = $derived(
    guruState.sekolahList.find((s) => s.id == guruState.activeSekolahId) || null
  );

  let activeTab = $state('tahfiz');

  let dataSantri = $state([]);
  let dataDiniyah = $state([]);
  let loadingData = $state(false);

  let selectedMetric = $state('hafalan');

  let avgHafalan = $derived(
    dataSantri.length
      ? Math.round(dataSantri.reduce((a, b) => a + b.hafalan, 0) / dataSantri.length)
      : 0
  );
  let avgLancar = $derived(
    dataSantri.length
      ? Math.round(dataSantri.reduce((a, b) => a + b.lancar, 0) / dataSantri.length)
      : 0
  );
  let avgMad = $derived(
    dataSantri.length
      ? Math.round(dataSantri.reduce((a, b) => a + b.mad, 0) / dataSantri.length)
      : 0
  );
  let avgTajwid = $derived(
    dataSantri.length
      ? Math.round(dataSantri.reduce((a, b) => a + b.tajwid, 0) / dataSantri.length)
      : 0
  );
  let avgMakhraj = $derived(
    dataSantri.length
      ? Math.round(dataSantri.reduce((a, b) => a + b.makhraj, 0) / dataSantri.length)
      : 0
  );

  let globalAverages = $derived({
    hafalan: avgHafalan,
    lancar: avgLancar,
    mad: avgMad,
    tajwid: avgTajwid,
    makhraj: avgMakhraj
  });

  let isTahfizV2 = $derived(
    guruState.activeSekolahId && 
    (guruState.aksesMap[guruState.activeSekolahId]?.includes('input_tahfiz_v2') || 
     guruState.aksesMap[guruState.activeSekolahId]?.includes('paket_alquran'))
  );

  let globalAveragesArray = $derived([
    !isTahfizV2 && { label: 'Hafalan (%)', value: avgHafalan },
    { label: 'Kelancaran', value: avgLancar },
    !isTahfizV2 && { label: 'Hukum Mad', value: avgMad },
    { label: 'Tajwid', value: avgTajwid },
    { label: 'Makhraj', value: avgMakhraj }
  ].filter(Boolean));

  let availableMetrics = $derived([
    !isTahfizV2 && { value: 'hafalan', label: 'Hafalan (%)' },
    { value: 'lancar', label: 'Kelancaran' },
    !isTahfizV2 && { value: 'mad', label: 'Hukum Mad' },
    { value: 'tajwid', label: 'Tajwid' },
    { value: 'makhraj', label: 'Makhraj' }
  ].filter(Boolean));

  $effect(() => {
    if (isTahfizV2 && (selectedMetric === 'hafalan' || selectedMetric === 'mad')) {
      selectedMetric = 'lancar';
    }
  });

  const metricConfig = {
    hafalan: { color: '#0ea5e9', label: 'Hafalan (%)' },
    lancar: { color: '#10b981', label: 'Kelancaran' },
    mad: { color: '#a855f7', label: 'Hukum Mad' },
    tajwid: { color: '#f59e0b', label: 'Tajwid' },
    makhraj: { color: '#f43f5e', label: 'Makhraj' }
  };

  function getRombelName(s) {
    let rName = '';
    if (s.tingkat && s.tingkat !== '-' && s.tingkat !== 'Belum ada Kelas') {
      rName += `Kelas ${s.tingkat}`;
    }
    if (s.rombel && s.rombel !== '-') {
      rName += rName ? ` ${s.rombel}` : s.rombel;
    }
    return rName || 'Belum ada Rombel';
  }

  let uniqueRombel = $derived([
    'Semua Rombel',
    ...new Set([
      ...dataSantri.map((s) => getRombelName(s)),
      ...dataDiniyah.map((d) => getRombelName(d))
    ])
  ]);

  let selectedRombel = $state('Semua Rombel');

  let filteredSantri = $derived(
    selectedRombel === 'Semua Rombel'
      ? dataSantri
      : dataSantri.filter((s) => getRombelName(s) === selectedRombel)
  );

  let rombelAverages = $derived(() => {
    const groups = {};
    for (const s of filteredSantri) {
      const r = getRombelName(s);

      if (!groups[r]) {
        groups[r] = {
          count: 0,
          hafalan: 0,
          lancar: 0,
          mad: 0,
          tajwid: 0,
          makhraj: 0,
          jumlahSurah: 0
        };
      }
      groups[r].count++;
      groups[r].hafalan += s.hafalan;
      groups[r].lancar += s.lancar;
      groups[r].mad += s.mad;
      groups[r].tajwid += s.tajwid;
      groups[r].makhraj += s.makhraj;
      groups[r].jumlahSurah += s.jumlahSurah;
    }

    return Object.keys(groups)
      .map((rombel) => {
        const g = groups[rombel];
        return {
          label: rombel,
          hafalan: g.hafalan / g.count,
          lancar: g.lancar / g.count,
          mad: g.mad / g.count,
          tajwid: g.tajwid / g.count,
          makhraj: g.makhraj / g.count,
          jumlahSurah: g.jumlahSurah / g.count
        };
      })
      .sort((a, b) => a.label.localeCompare(b.label));
  });

  let detailData = $derived(
    selectedRombel === 'Semua Rombel'
      ? rombelAverages().map((r) => ({
          label: r.label,
          value: r[selectedMetric]
        }))
      : filteredSantri.map((s) => ({
          label: s.nama,
          value: s[selectedMetric]
        }))
  );

  let detailColor = $derived(metricConfig[selectedMetric].color);

  let surahData = $derived(
    selectedRombel === 'Semua Rombel'
      ? rombelAverages().map((r) => ({
          label: r.label,
          value: r.jumlahSurah
        }))
      : filteredSantri.map((s) => ({
          label: s.nama,
          value: s.jumlahSurah
        }))
  );

  // --- Logika Statistik Diniyah ---
  let uniqueMapel = $derived([
    'Semua Mapel',
    ...new Set(dataDiniyah.map((d) => d.mapel).filter((m) => m && m !== 'Tanpa Mapel'))
  ]);
  let selectedMapel = $state('Semua Mapel');

  let filteredDiniyah = $derived(
    dataDiniyah.filter((d) => {
      const rName = getRombelName(d);
      const matchRombel = selectedRombel === 'Semua Rombel' || rName === selectedRombel;
      const matchMapel = selectedMapel === 'Semua Mapel' || d.mapel === selectedMapel;
      return matchRombel && matchMapel;
    })
  );

  let diniyahChartData = $derived(() => {
    const groups = {};
    for (const d of filteredDiniyah) {
      const key = selectedRombel === 'Semua Rombel' ? getRombelName(d) : d.nama;
      if (!groups[key]) {
        groups[key] = { count: 0, total: 0 };
      }
      groups[key].count++;
      groups[key].total += d.nilai;
    }

    return Object.keys(groups)
      .map((k) => ({
        label: k,
        value: groups[k].total / groups[k].count
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  });

  import { untrack } from 'svelte';
  $effect(() => {
    const sekolahId = guruState.activeSekolahId;
    const taId = guruState.activeTahunId;
    const semId = guruState.activeSemesterId;
    if (sekolahId && taId && semId) {
      untrack(() => loadDashboardData(taId, semId));
    }
  });

  // OFFLINE-FIRST: Brankas Dinamis
  $effect(() => {
    if (typeof window !== 'undefined' && selectedSekolah && (dataSantri.length > 0 || dataDiniyah.length > 0)) {
       const cacheKey = `siakad_offline_dashboard_${selectedSekolah.id}_${guruState.activeTahunId}_${guruState.activeSemesterId}`;
       localStorage.setItem(cacheKey, JSON.stringify({ dataSantri, dataDiniyah }));
    }
  });

  async function loadDashboardData(taId = guruState.activeTahunId, semId = guruState.activeSemesterId) {
    if (!selectedSekolah) return;
    
    // 0. OFFLINE-FIRST UX (Stale-While-Revalidate)
    if (typeof window !== 'undefined') {
       const cacheKey = `siakad_offline_dashboard_${selectedSekolah.id}_${taId}_${semId}`;
       const cached = localStorage.getItem(cacheKey);
       if (cached) {
         try {
           const data = JSON.parse(cached);
           if (data.dataSantri?.length > 0 || data.dataDiniyah?.length > 0) {
             dataSantri = data.dataSantri || [];
             dataDiniyah = data.dataDiniyah || [];
             loadingData = false;
           }
         } catch(e) {}
       }
    }

    if (dataSantri.length === 0 && dataDiniyah.length === 0) {
      loadingData = true;
    }
    
    try {
      const params = new URLSearchParams();
      if (taId) params.append('tahun_ajaran_id', taId);
      if (semId) params.append('semester_id', semId);
      const queryStr = params.toString() ? `?${params.toString()}` : '';

      const resTahfiz = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/dashboard/tahfiz${queryStr}`,
        { credentials: 'include' }
      );
      if (resTahfiz.ok) {
        dataSantri = (await resTahfiz.json()) || [];
      }

      const resDiniyah = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/dashboard/diniyah${queryStr}`,
        { credentials: 'include' }
      );
      if (resDiniyah.ok) {
        dataDiniyah = (await resDiniyah.json()) || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loadingData = false;
    }
  }
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

    <!-- Tab Navigation -->
    <div class="tabs">
      <button
        class="tab-btn"
        class:active={activeTab === 'tahfiz'}
        onclick={() => (activeTab = 'tahfiz')}
      >
        <Icon name="book" size={18} /> Statistik Tahfiz
      </button>
      <button
        class="tab-btn"
        class:active={activeTab === 'diniyah'}
        onclick={() => (activeTab = 'diniyah')}
      >
        <Icon name="bookOpen" size={18} /> Statistik Diniyah
      </button>
    </div>

    {#if loadingData}
      <div class="loading-state">Memuat data dashboard...</div>
    {:else}
      {#if activeTab === 'tahfiz'}
        {#if dataSantri.length === 0}
          <div class="empty-state">Belum ada data nilai tahfiz untuk sekolah ini.</div>
        {:else}
          <!-- Summary Cards -->
          <div class="summary-grid">
            <div class="summary-card">
              <div class="icon-box" style="background: #eff6ff; color: #3b82f6;">
                <Icon name="users" size={20} />
              </div>
              <div>
                <p>Jumlah Santri Dievaluasi</p>
                <h3>{dataSantri.length}</h3>
              </div>
            </div>
          </div>

          <!-- Global Chart -->
          <div class="global-chart-box">
            <div class="blur-blob top-right"></div>
            <div class="blur-blob bottom-left"></div>

            <div class="chart-header">
              <h3>Rata-rata Global Kelas (Tahfiz)</h3>
              <p>Pencapaian akumulatif dari {dataSantri.length} santri</p>
            </div>

            <BarChart 
              data={globalAveragesArray} 
              color="#0ea5e9"
              max={100}
              labelFormat={(v) => Math.round(v)}
            />
          </div>

          <!-- Detail Charts -->
          <div class="charts-grid">
            <!-- Chart Kualitas per Santri -->
            <div class="chart-card">
              <div class="chart-header-row">
                <div>
                  <h3>
                    {selectedRombel === 'Semua Rombel'
                      ? 'Statistik Per Rombel'
                      : `Statistik Siswa - ${selectedRombel}`}
                  </h3>
                  <p>
                    {selectedRombel === 'Semua Rombel'
                      ? 'Perbandingan rata-rata nilai antar rombel'
                      : 'Perbandingan nilai masing-masing siswa'}
                  </p>
                </div>
                <div class="filter-group">
                  <select bind:value={selectedRombel} class="metric-select">
                    {#each uniqueRombel as rombel}
                      <option value={rombel}>{rombel}</option>
                    {/each}
                  </select>
                  <select bind:value={selectedMetric} class="metric-select">
                    {#each availableMetrics as m}
                      <option value={m.value}>{m.label}</option>
                    {/each}
                  </select>
                </div>
              </div>
              <BarChart
                data={detailData}
                color={detailColor}
                max={100}
                labelFormat={(v) => Math.round(v) + '%'}
              />
            </div>

            <!-- Chart Jumlah Surah -->
            <div class="chart-card">
              <h3>Status Jumlah Surah</h3>
              <p>
                {selectedRombel === 'Semua Rombel'
                  ? 'Rata-rata hafalan surah per rombel'
                  : 'Total hafalan surah siswa'}
              </p>
              <BarChart
                data={surahData}
                color="#4f46e5"
                max={37}
                labelFormat={(v) => Math.round(v) + ' Surah'}
              />
            </div>
          </div>
        {/if}
      {:else if activeTab === 'diniyah'}
        {#if dataDiniyah.length === 0}
          <div class="empty-state">Belum ada data nilai diniyah untuk sekolah ini.</div>
        {:else}
          <div class="chart-card diniyah-chart-card">
            <div class="chart-header-row">
              <div>
                <h3>
                  {selectedRombel === 'Semua Rombel'
                    ? 'Statistik Diniyah Per Rombel'
                    : `Statistik Diniyah Siswa - ${selectedRombel}`}
                </h3>
                <p>
                  {selectedRombel === 'Semua Rombel'
                    ? 'Rata-rata nilai rapor diniyah antar rombel'
                    : 'Rata-rata nilai rapor diniyah masing-masing siswa'}
                </p>
              </div>
              <div class="filter-group">
                <select bind:value={selectedRombel} class="metric-select">
                  {#each uniqueRombel as rombel}
                    <option value={rombel}>{rombel}</option>
                  {/each}
                </select>
                <select bind:value={selectedMapel} class="metric-select">
                  {#each uniqueMapel as mapel}
                    <option value={mapel}>{mapel}</option>
                  {/each}
                </select>
              </div>
            </div>

            <BarChart
              data={diniyahChartData()}
              color="#10b981"
              max={100}
              labelFormat={(v) => Math.round(v)}
            />
          </div>
        {/if}
      {/if}
    {/if}
  {/if}
</div>

<!-- TAMPILAN MOBILE: BERANDA GURU -->
<div class="mobile-app-ui sheet-page" style="padding-top: 0 !important; margin-bottom: 80px;">
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
    
    {#if selectedSekolah && !loadingData}
    <div class="stat-card">
      <div class="stat-item">
        <div class="stat-num">{dataSantri.length}</div>
        <div class="stat-label">Santri Tahfiz</div>
      </div>
      <div class="stat-item">
        <div class="stat-num">{dataDiniyah.length}</div>
        <div class="stat-label">Santri Diniyah</div>
      </div>
    </div>
    {/if}
  </div>

  {#if selectedSekolah}
    <!-- WIDGET PRESENSI (Langsung Tampil di Beranda) -->
    <div style="margin: 0 1rem;">
      <PresensiWidget />
    </div>

    {#if loadingData}
      <div class="empty-state" style="margin-top: 1rem; border: none;">
        Memuat data dashboard...
      </div>
    {:else}
      <div class="mobile-cards-view" style="padding-bottom: 0;">
        <div class="surah-card" style="margin-top: 0;">
          <div class="surah-top" style="border-bottom: 1px solid var(--border);">
            <div class="surah-no" style="background: #e0f2fe; color: #0284c7;"><Icon name="book" size={14}/></div>
            <div class="surah-name">Statistik Tahfiz (Global)</div>
          </div>
          <div style="padding: 12px 14px;">
            {#if dataSantri.length === 0}
              <div style="font-size: 13px; color: var(--muted); text-align: center;">Belum ada data nilai tahfiz.</div>
            {:else}
              {#each globalAveragesArray as avg}
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                  <span style="font-size: 13px; font-weight: 600; color: var(--muted);">{avg.label}</span>
                  <span style="font-size: 14px; font-weight: 800; color: var(--text);">{Math.round(avg.value)}</span>
                </div>
                <div style="width: 100%; background: var(--border); height: 6px; border-radius: 4px; margin-bottom: 12px; overflow: hidden;">
                  <div style="height: 100%; background: var(--primary); border-radius: 4px; width: {Math.round(avg.value)}%;"></div>
                </div>
              {/each}
            {/if}
          </div>
        </div>

        <div class="surah-card">
          <div class="surah-top" style="border-bottom: 1px solid var(--border);">
            <div class="surah-no" style="background: #dcfce7; color: #16a34a;"><Icon name="bookOpen" size={14}/></div>
            <div class="surah-name">Statistik Diniyah</div>
          </div>
          <div style="padding: 12px 14px;">
            {#if dataDiniyah.length === 0}
              <div style="font-size: 13px; color: var(--muted); text-align: center;">Belum ada data nilai diniyah.</div>
            {:else}
              {#each diniyahChartData() as dChart (dChart.label)}
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                  <span style="font-size: 13px; font-weight: 600; color: var(--muted);">{dChart.label}</span>
                  <span style="font-size: 14px; font-weight: 800; color: var(--text);">{Math.round(dChart.value)}</span>
                </div>
                <div style="width: 100%; background: var(--border); height: 6px; border-radius: 4px; margin-bottom: 12px; overflow: hidden;">
                  <div style="height: 100%; background: #10b981; border-radius: 4px; width: {Math.round(dChart.value)}%;"></div>
                </div>
              {/each}
            {/if}
          </div>
        </div>
      </div>
    {/if}
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

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .summary-card {
    background: white;
    padding: 1.25rem;
    border-radius: 1rem;
    border: 1px solid #f1f5f9;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .icon-box {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .summary-card p {
    margin: 0 0 0.25rem 0;
    font-size: 0.85rem;
    color: #64748b;
    font-weight: 500;
  }

  .summary-card h3 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: #0f172a;
    font-family: 'Outfit', sans-serif;
  }

  .global-chart-box {
    background: #ffffff;
    border-radius: 1.5rem;
    padding: 2rem;
    margin-bottom: 2rem;
    position: relative;
    overflow: hidden;
    border: 1px solid #f1f5f9;
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.05);
  }

  .blur-blob {
    position: absolute;
    width: 250px;
    height: 250px;
    border-radius: 50%;
    filter: blur(80px);
    pointer-events: none;
    opacity: 0.1;
  }

  .top-right {
    top: -50px;
    right: -50px;
    background: #0ea5e9;
  }

  .bottom-left {
    bottom: -50px;
    left: -50px;
    background: #10b981;
  }

  .chart-header {
    text-align: center;
    margin-bottom: 1rem;
    position: relative;
    z-index: 2;
  }

  .chart-header h3 {
    color: #0f172a;
    margin: 0 0 0.25rem 0;
    font-size: 1.5rem;
    font-family: 'Outfit', sans-serif;
  }

  .chart-header p {
    color: #64748b;
    margin: 0;
    font-size: 0.9rem;
  }

  .charts-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .chart-card {
    background: white;
    border-radius: 1rem;
    padding: 1.5rem;
    border: 1px solid #f1f5f9;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    display: flex;
    flex-direction: column;
  }

  .chart-card h3 {
    margin: 0 0 0.25rem 0;
    font-size: 1.25rem;
    font-family: 'Outfit', sans-serif;
    color: #0f172a;
  }

  .chart-card p {
    margin: 0 0 1.5rem 0;
    font-size: 0.85rem;
    color: #64748b;
  }

  .chart-header-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .chart-header-row p {
    margin-bottom: 0;
  }

  .metric-select {
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    color: #334155;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    font-size: 0.85rem;
    font-weight: 500;
    outline: none;
    cursor: pointer;
  }

  .metric-select:focus {
    border-color: #0ea5e9;
    box-shadow: 0 0 0 2px rgba(14, 165, 233, 0.2);
  }

  .filter-group {
    display: flex;
    gap: 0.75rem;
    align-items: center;
  }

  /* Tabs */
  .tabs {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
    border-bottom: 2px solid #e2e8f0;
    padding-bottom: 0;
  }
  .tab-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: none;
    border: none;
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    font-weight: 600;
    color: #64748b;
    cursor: pointer;
    border-bottom: 3px solid transparent;
    margin-bottom: -2px;
    transition: all 0.2s;
  }
  .tab-btn:hover {
    color: #0f172a;
  }
  .tab-btn.active {
    color: #0ea5e9;
    border-bottom-color: #0ea5e9;
  }

  .loading-state,
  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #64748b;
    background: white;
    border-radius: 1rem;
    border: 1px dashed #cbd5e1;
  }

  /* Vanilla Svelte Bar Chart */
  .diniyah-chart-card {
    background: white;
    padding: 2rem;
  }
  .vanilla-bar-chart {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    margin-top: 1.5rem;
  }
  .bar-row {
    display: flex;
    align-items: center;
    gap: 1rem;
  }
  .bar-label {
    width: 150px;
    min-width: 150px;
    font-weight: 500;
    color: #334155;
    text-align: right;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .bar-track {
    flex: 1;
    height: 32px;
    background: #f1f5f9;
    border-radius: 6px;
    position: relative;
    overflow: hidden;
  }
  .bar-fill {
    height: 100%;
    background: linear-gradient(90deg, #10b981 0%, #34d399 100%);
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding-right: 0.75rem;
    transition: width 1s cubic-bezier(0.4, 0, 0.2, 1);
    min-width: 40px;
  }
  .bar-value {
    color: white;
    font-weight: 700;
    font-size: 0.9rem;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  }
  @media (max-width: 768px) {
    .beranda-guru {
      padding: 0 0.5rem;
    }
    .global-chart-box, .chart-card, .diniyah-chart-card {
      padding: 1rem;
    }
    .chart-header-row {
      flex-direction: column;
      gap: 1rem;
    }
    .filter-group {
      width: 100%;
      flex-direction: column;
    }
    .metric-select {
      width: 100%;
    }
    .tabs {
      overflow-x: auto;
      white-space: nowrap;
      padding-bottom: 0.5rem;
    }
    .tab-btn {
      padding: 0.5rem 1rem;
      font-size: 0.9rem;
    }
  }
</style>
