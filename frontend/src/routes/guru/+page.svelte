<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import RadialChart from '$lib/components/charts/RadialChart.svelte';
  import BarChart from '$lib/components/charts/BarChart.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

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

  $effect(() => {
    if (selectedSekolah) {
      loadDashboardData();
    }
  });

  async function loadDashboardData() {
    if (!selectedSekolah) return;
    loadingData = true;
    try {
      const resTahfiz = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/dashboard/tahfiz`,
        { credentials: 'include' }
      );
      if (resTahfiz.ok) {
        dataSantri = (await resTahfiz.json()) || [];
      }

      const resDiniyah = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/dashboard/diniyah`,
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

<div class="beranda-guru">
  {#if !selectedSekolah}
    <div class="empty-state">
      <Icon name="grid" size={48} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas untuk melihat dashboard.</p>
    </div>
  {:else}
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

            <RadialChart averages={globalAverages} />
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
                    <option value="hafalan">Hafalan (%)</option>
                    <option value="lancar">Kelancaran</option>
                    <option value="mad">Hukum Mad</option>
                    <option value="tajwid">Tajwid</option>
                    <option value="makhraj">Makhraj</option>
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

<style>
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
</style>
