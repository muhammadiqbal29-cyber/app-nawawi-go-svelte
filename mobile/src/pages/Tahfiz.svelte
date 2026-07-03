<script lang="ts">
    import { onMount } from 'svelte';
    import { router } from 'tinro';
    import { getHalaqahKu, getAnggotaHalaqah } from '../services/endpoints';
    import Icon from '../components/Icon.svelte';
    
    let isDailyViewOpen = $state(false);
    let halaqahList = $state<any[]>([]);
    let selectedHalaqahId = $state<number | null>(null);
    let santriList = $state<any[]>([]);
    let selectedSantri = $state<any>(null);
    let searchQuery = $state('');

    // Data Kualitas Bacaan
    const tahfizMetrics = [
        { name: 'Tajwid', score: 90, color: '#10B981' },
        { name: 'Kelancaran', score: 85, color: '#3B82F6' },
        { name: 'Makharijul Huruf', score: 88, color: '#F59E0B' },
        { name: 'Ghunnah', score: 92, color: '#8B5CF6' },
        { name: 'Mad (Panjang)', score: 75, color: '#EF4444' }
    ];

    let searchResults = $derived(() => {
        if (!searchQuery) return [];
        const lowerQ = searchQuery.toLowerCase();
        return santriList.filter(s => s.siswa?.nama_lengkap?.toLowerCase().includes(lowerQ));
    });

    function toggleDailyProgress() {
        isDailyViewOpen = !isDailyViewOpen;
    }

    async function loadData() {
        try {
            const hList = await getHalaqahKu();
            halaqahList = hList || [];
            if (halaqahList.length > 0) {
                selectedHalaqahId = halaqahList[0].id;
                await onHalaqahChange();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function onHalaqahChange() {
        if (!selectedHalaqahId) return;
        try {
            const sList = await getAnggotaHalaqah(selectedHalaqahId);
            santriList = sList || [];
            if (santriList.length > 0) {
                selectedSantri = santriList[0];
                searchQuery = '';
            } else {
                selectedSantri = null;
            }
        } catch (e) {
            console.error(e);
        }
    }

    function selectSantri(santri: any) {
        selectedSantri = santri;
        searchQuery = '';
    }

    onMount(() => {
        loadData();

        setTimeout(() => {
            const juzCircle = document.getElementById('juzCircle');
            if (juzCircle) {
                const totalJuz = 30;
                const hafalJuz = 5;
                const percentage = (hafalJuz / totalJuz) * 100;
                const offset = 100 - percentage; 
                juzCircle.style.strokeDashoffset = offset.toString();
            }
        }, 300);
    });
</script>

<div class="t-wrapper">
    <div class="t-container">

        <!-- Header -->
        <div class="t-header">
            <button onclick={() => router.goto('/menu_perkembangan')} class="t-back-btn">
                <Icon name="arrow-left" size="14px" /> Kembali
            </button>
            
            <div class="t-header-controls">
                <!-- Dropdown -->
                <div class="t-select-wrapper">
                    <select bind:value={selectedHalaqahId} onchange={onHalaqahChange} class="t-select">
                        {#each halaqahList as h}
                            <option value={h.id}>{h.nama_halaqah}</option>
                        {/each}
                    </select>
                    <div class="t-select-icon">
                        <Icon name="chevron-down" size="12px" />
                    </div>
                </div>

                <!-- Search -->
                <div class="t-search-wrapper">
                    <div class="t-search-icon">
                        <Icon name="search" size="14px" />
                    </div>
                    <input type="text" bind:value={searchQuery} placeholder="Cari santri..." class="t-search-input">
                    {#if searchResults().length > 0 && searchQuery}
                        <div class="t-search-dropdown">
                            {#each searchResults() as santri}
                                <button onclick={() => selectSantri(santri)} class="t-search-item">
                                    {santri.siswa?.nama_lengkap}
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            </div>

            {#if selectedSantri}
                <h1 class="t-student-name">{selectedSantri.siswa?.nama_lengkap}</h1>
                <p class="t-student-nis">NIS: {selectedSantri.siswa?.nis}</p>
                <div class="t-badge-target">TARGET LULUS: 3 JUZ / TAHUN</div>
            {:else}
                <h1 class="t-empty-title">Pilih Santri</h1>
                <p class="t-empty-subtitle">Silakan cari nama santri di atas</p>
            {/if}
        </div>

        {#if selectedSantri}
        <div class="t-main-view" class:hidden={isDailyViewOpen}>
            
            <!-- Kiri: Circular Progress & Radar -->
            <div class="t-left-panel">
                <h2 class="t-panel-title">Total Pencapaian</h2>
                <p class="t-panel-subtitle">Progress hafalan mutqin (lancar)</p>
                
                <div class="t-circle-wrapper">
                    <svg class="t-circle-svg" viewBox="0 0 36 36">
                        <path class="t-circle-bg" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                        <path id="juzCircle" class="t-circle-fg animate-circle" stroke-dasharray="100, 100" stroke-dashoffset="100" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                    </svg>
                    <div class="t-circle-text">
                        <span class="t-circle-num">5</span>
                        <span class="t-circle-lbl">dari 30 Juz</span>
                    </div>
                </div>

                <div class="t-divider"></div>

                <h2 class="t-panel-title" style="margin-top: 1rem;">Kualitas Bacaan</h2>
                <div class="t-bar-chart-container">
                    {#each tahfizMetrics as metric}
                        <div class="t-bar-metric-item">
                            <div class="t-bar-metric-header">
                                <span class="t-bar-metric-name">{metric.name}</span>
                                <span class="t-bar-metric-score" style="color: {metric.color}">{metric.score}</span>
                            </div>
                            <div class="t-progress-bar-bg t-bar-metric-bg">
                                <div class="t-progress-bar-fill-blue animate-progress" style="width: {metric.score}%; background-color: {metric.color}"></div>
                            </div>
                        </div>
                    {/each}
                </div>
            </div>

            <!-- Kanan: Ziyadah & Murojaah -->
            <div class="t-right-panel">
                <div class="t-panel-header">
                    <h2 class="t-panel-title-lg">Aktivitas Pekan Ini</h2>
                    <button onclick={toggleDailyProgress} class="t-btn-history">Riwayat Harian</button>
                </div>
                
                <!-- Ziyadah Card -->
                <div class="t-card t-card-blue">
                    <div class="t-card-bg-icon"></div>
                    <div class="t-card-header">
                        <div class="t-card-title-wrap">
                            <div class="t-card-icon-blue">Z</div>
                            <div>
                                <h3>Ziyadah</h3>
                                <p>Hafalan Baru</p>
                            </div>
                        </div>
                        <span class="t-card-badge-blue">Juz 29</span>
                    </div>
                    
                    <div class="t-progress-item">
                        <div class="t-progress-label">
                            <span>Surah Al-Mulk (Ayat 1-30)</span>
                            <span class="t-bold">100%</span>
                        </div>
                        <div class="t-progress-bar-bg">
                            <div class="t-progress-bar-fill-blue animate-progress" style="width: 100%;"></div>
                        </div>
                    </div>

                    <div class="t-progress-item">
                        <div class="t-progress-label">
                            <span>Surah Al-Qalam (Ayat 1-25)</span>
                            <span class="t-bold">48%</span>
                        </div>
                        <div class="t-progress-bar-bg">
                            <div class="t-progress-bar-fill-blue-light animate-progress" style="width: 48%;"></div>
                        </div>
                    </div>
                </div>

                <!-- Murojaah Card -->
                <div class="t-card t-card-orange">
                    <div class="t-card-bg-icon"></div>
                    <div class="t-card-header">
                        <div class="t-card-title-wrap">
                            <div class="t-card-icon-orange">M</div>
                            <div>
                                <h3>Murojaah</h3>
                                <p>Pengulangan Hafalan Lama</p>
                            </div>
                        </div>
                        <span class="t-card-badge-orange">Juz 30</span>
                    </div>
                    
                    <div class="t-stats-grid">
                        <div class="t-stat-box">
                            <div class="t-stat-val t-text-orange">15</div>
                            <div class="t-stat-lbl">Surah Diulang</div>
                        </div>
                        <div class="t-stat-box">
                            <div class="t-stat-val t-text-green">Sangat Lancar</div>
                            <div class="t-stat-lbl">Rata-rata Status</div>
                        </div>
                    </div>

                    <div class="t-murojaah-bar-wrap">
                        <div class="t-murojaah-lbl">
                            <span>Kekuatan Ingatan (Mutqin)</span>
                            <span class="t-text-orange">85% Kuat</span>
                        </div>
                        <div class="t-progress-bar-bg t-flex-bar">
                            <div class="t-bar-green animate-progress" style="width: 50%;"></div>
                            <div class="t-bar-yellow animate-progress" style="width: 35%;"></div>
                            <div class="t-bar-red animate-progress" style="width: 15%;"></div>
                        </div>
                        <div class="t-murojaah-legend">
                            <span class="t-text-green">Sangat Lancar</span>
                            <span class="t-text-yellow">Lancar</span>
                            <span class="t-text-red">Lupa/Tersendat</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Daily View -->
        <div class="t-daily-view" class:hidden={!isDailyViewOpen}>
            <div class="t-panel-header">
                <h2 class="t-panel-title-lg">Riwayat Setoran Harian</h2>
                <button onclick={toggleDailyProgress} class="t-btn-back-outline">Kembali</button>
            </div>
            
            <div class="t-timeline">
                
                <div class="t-timeline-item">
                    <div class="t-timeline-dot t-dot-green"></div>
                    <div class="t-timeline-date">Hari ini (Jumat, 20 Juni)</div>
                    <div class="t-timeline-box t-box-blue">
                        <span class="t-badge-sm-blue">Ziyadah</span>
                        <span class="t-box-text">Surah Al-Mulk ayat 15-30 <span class="t-text-green t-bold">(Sangat Lancar)</span></span>
                    </div>
                </div>

                <div class="t-timeline-item">
                    <div class="t-timeline-dot t-dot-orange"></div>
                    <div class="t-timeline-date">Kamis, 19 Juni</div>
                    <div class="t-timeline-box t-box-orange">
                        <span class="t-badge-sm-orange">Murojaah</span>
                        <span class="t-box-text">Juz 30 (An-Naba s.d At-Takwir) <span class="t-text-green t-bold">(Sangat Lancar)</span></span>
                    </div>
                    <div class="t-timeline-box t-box-orange">
                        <span class="t-badge-sm-orange">Murojaah</span>
                        <span class="t-box-text">Surah Al-Mursalat <span class="t-text-yellow t-bold">(Lancar, 1x salah)</span></span>
                    </div>
                </div>

                <div class="t-timeline-item">
                    <div class="t-timeline-dot t-dot-green"></div>
                    <div class="t-timeline-date">Rabu, 18 Juni</div>
                    <div class="t-timeline-box t-box-blue">
                        <span class="t-badge-sm-blue">Ziyadah</span>
                        <span class="t-box-text">Surah Al-Mulk ayat 1-14 <span class="t-text-green t-bold">(Sangat Lancar)</span></span>
                    </div>
                    <div class="t-timeline-box t-box-orange">
                        <span class="t-badge-sm-orange">Murojaah</span>
                        <span class="t-box-text">Surah Al-Qalam ayat 1-25 <span class="t-text-red t-bold">(Banyak Lupa)</span></span>
                    </div>
                </div>
                
            </div>
        </div>
        {/if}
    </div>
</div>

<style>
    /* Vanilla CSS to replace Tailwind */
    .hidden { display: none !important; }
    .t-bold { font-weight: 700; }
    
    .t-wrapper {
        display: flex; align-items: center; justify-content: center;
        min-height: 100vh; background-color: #f3f4f6; padding: 0;
    }
    .t-container {
        background-color: white; width: 100%; min-height: 100vh;
        position: relative; overflow: hidden; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1);
    }
    @media (min-width: 768px) {
        .t-wrapper { padding: 3rem 1.5rem; }
        .t-container { max-width: 56rem; min-height: 0; border-radius: 1.5rem; border: 1px solid #e5e7eb; }
    }

    /* Header */
    .t-header {
        background-color: #06b6d4; color: white; padding: 1.5rem; text-align: center;
        position: relative; z-index: 10; border-bottom: 4px solid #0891b2; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1);
    }
    @media (min-width: 768px) { .t-header { padding: 2rem; } }
    
    .t-back-btn {
        position: absolute; top: 1.5rem; left: 1.5rem; color: rgba(255,255,255,0.8); background: transparent; border: none;
        cursor: pointer; font-size: 0.875rem; font-weight: 600; display: flex; align-items: center; gap: 0.5rem; transition: 0.2s;
    }
    .t-back-btn:hover { color: white; }
    
    .t-header-controls {
        display: flex; flex-direction: column; gap: 0.75rem; justify-content: center; align-items: center;
        margin-top: 2.5rem; margin-bottom: 1rem;
    }
    @media (min-width: 768px) { .t-header-controls { flex-direction: row; } }
    
    .t-select-wrapper, .t-search-wrapper { position: relative; width: 100%; max-width: 20rem; text-align: left; }
    
    .t-select, .t-search-input {
        width: 100%; background: rgba(8, 145, 178, 0.5); border: 1px solid #22d3ee; color: white;
        font-size: 0.875rem; font-weight: 600; border-radius: 0.75rem; padding: 0.625rem; outline: none; box-sizing: border-box;
    }
    .t-select { appearance: none; padding-right: 2rem; }
    .t-search-input { padding-left: 2.5rem; }
    .t-search-input::placeholder { color: #a5f3fc; }
    .t-select:focus, .t-search-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.5); }
    .t-select option { color: #111827; }
    
    .t-select-icon { position: absolute; right: 0.75rem; top: 0; bottom: 0; display: flex; align-items: center; pointer-events: none; color: white; }
    .t-search-icon { position: absolute; left: 0.75rem; top: 0; bottom: 0; display: flex; align-items: center; pointer-events: none; color: #a5f3fc; }
    
    .t-search-dropdown {
        position: absolute; z-index: 50; width: 100%; margin-top: 0.25rem; background: white; border: 1px solid #e5e7eb;
        border-radius: 0.75rem; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1); max-height: 15rem; overflow-y: auto;
    }
    .t-search-item {
        width: 100%; text-align: left; padding: 0.5rem 1rem; font-size: 0.875rem; font-weight: 500; color: #374151;
        background: transparent; border: none; border-bottom: 1px solid #f3f4f6; cursor: pointer; transition: 0.2s;
    }
    .t-search-item:last-child { border-bottom: none; }
    .t-search-item:hover { background: #ecfeff; }
    
    .t-student-name { font-size: 1.5rem; font-weight: 900; letter-spacing: 0.025em; margin: 1rem 0 0.5rem; text-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    @media (min-width: 768px) { .t-student-name { font-size: 1.875rem; margin-top: 0; } }
    .t-student-nis { font-size: 0.875rem; font-weight: 600; color: rgba(255,255,255,0.9); margin: 0 0 0.75rem; }
    @media (min-width: 768px) { .t-student-nis { font-size: 1rem; } }
    .t-badge-target {
        display: inline-block; background: rgba(255,255,255,0.2); padding: 0.375rem 1rem; border-radius: 9999px;
        font-size: 0.75rem; font-weight: 700; letter-spacing: 0.05em; backdrop-filter: blur(4px); box-shadow: 0 1px 2px rgba(0,0,0,0.05);
    }
    .t-empty-title { font-size: 1.25rem; font-weight: 700; letter-spacing: 0.025em; margin: 1.5rem 0 0.5rem; text-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .t-empty-subtitle { font-size: 0.875rem; font-weight: 500; color: rgba(255,255,255,0.8); margin: 0; }
    
    /* Layout */
    .t-main-view { display: flex; flex-direction: column; }
    @media (min-width: 768px) { .t-main-view { flex-direction: row; } }
    
    /* Left Panel */
    .t-left-panel {
        width: 100%; padding: 1.5rem; display: flex; flex-direction: column; align-items: center; background-color: rgba(236, 253, 245, 0.4); box-sizing: border-box;
    }
    @media (min-width: 768px) { .t-left-panel { width: 50%; padding: 2rem; } }
    .t-panel-title { font-size: 1.125rem; font-weight: 700; color: #1f2937; margin: 0 0 0.5rem; }
    .t-panel-subtitle { font-size: 0.75rem; font-weight: 600; color: #6b7280; margin: 0 0 1.5rem; }
    
    .t-circle-wrapper { position: relative; width: 12rem; height: 12rem; margin-bottom: 2rem; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.05)); }
    .t-circle-svg { width: 100%; height: 100%; transform: rotate(-90deg); }
    .t-circle-bg { stroke: #e5e7eb; stroke-width: 3; fill: none; }
    .t-circle-fg { stroke: #10b981; stroke-width: 3; fill: none; stroke-linecap: round; }
    .t-circle-text { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; }
    .t-circle-num { font-size: 3rem; font-weight: 900; color: #1f2937; letter-spacing: -0.05em; line-height: 1; }
    .t-circle-lbl { font-size: 0.875rem; font-weight: 700; color: #6b7280; margin-top: 0.25rem; }
    
    .t-divider { width: 100%; border-top: 1px solid #e5e7eb; margin: 1rem 0; }
    
    /* Bar Chart Metrics */
    .t-bar-chart-container { width: 100%; display: flex; flex-direction: column; gap: 1rem; margin-top: 0.5rem; }
    .t-bar-metric-item { width: 100%; }
    .t-bar-metric-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.375rem; }
    .t-bar-metric-name { font-size: 0.875rem; font-weight: 600; color: #374151; }
    .t-bar-metric-score { font-size: 0.875rem; font-weight: 800; }
    .t-bar-metric-bg { background-color: #f3f4f6; height: 0.625rem; border-radius: 9999px; }

    /* Right Panel */
    .t-right-panel {
        width: 100%; padding: 1.5rem; background-color: white; flex-grow: 1; box-sizing: border-box;
    }
    @media (min-width: 768px) { .t-right-panel { width: 50%; padding: 2rem; border-left: 1px solid #f3f4f6; } }
    .t-panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
    .t-panel-title-lg { font-size: 1.125rem; font-weight: 700; color: #1f2937; margin: 0; }
    @media (min-width: 768px) { .t-panel-title-lg { font-size: 1.25rem; } }
    .t-btn-history {
        font-size: 0.75rem; background: #d1fae5; color: #047857; padding: 0.375rem 1rem; border-radius: 9999px;
        font-weight: 700; border: none; cursor: pointer; box-shadow: 0 1px 2px rgba(0,0,0,0.05); transition: 0.2s;
    }
    .t-btn-history:hover { background: #a7f3d0; }
    
    /* Cards */
    .t-card { padding: 1.25rem; border-radius: 1rem; margin-bottom: 1.5rem; box-shadow: 0 1px 2px rgba(0,0,0,0.05); position: relative; overflow: hidden; }
    .t-card-bg-icon { position: absolute; top: 0; right: 0; width: 6rem; height: 6rem; border-bottom-left-radius: 9999px; opacity: 0.05; }
    .t-card-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1rem; position: relative; z-index: 10; }
    .t-card-title-wrap { display: flex; align-items: center; gap: 0.75rem; }
    .t-card-title-wrap h3 { font-weight: 700; margin: 0; line-height: 1.25; font-size: 1rem; }
    .t-card-title-wrap p { font-size: 0.75rem; font-weight: 600; margin: 0; }
    
    .t-card-blue { background: #eff6ff; border: 1px solid #dbeafe; }
    .t-card-blue .t-card-bg-icon { background: #3b82f6; }
    .t-card-icon-blue { width: 2.25rem; height: 2.25rem; border-radius: 9999px; background: #3b82f6; color: white; display: flex; align-items: center; justify-content: center; font-weight: 700; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }
    .t-card-blue h3 { color: #1e3a8a; }
    .t-card-blue p { color: #2563eb; }
    .t-card-badge-blue { font-size: 0.75rem; font-weight: 700; color: #1d4ed8; background: rgba(219,234,254,0.8); padding: 0.375rem 0.75rem; border-radius: 0.5rem; border: 1px solid #bfdbfe; }
    
    .t-card-orange { background: #fff7ed; border: 1px solid #ffedd5; }
    .t-card-orange .t-card-bg-icon { background: #f97316; }
    .t-card-icon-orange { width: 2.25rem; height: 2.25rem; border-radius: 9999px; background: #f97316; color: white; display: flex; align-items: center; justify-content: center; font-weight: 700; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }
    .t-card-orange h3 { color: #7c2d12; }
    .t-card-orange p { color: #ea580c; }
    .t-card-badge-orange { font-size: 0.75rem; font-weight: 700; color: #c2410c; background: rgba(255,237,213,0.8); padding: 0.375rem 0.75rem; border-radius: 0.5rem; border: 1px solid #fed7aa; }
    
    .t-progress-item { position: relative; z-index: 10; margin-bottom: 1rem; }
    .t-progress-item:last-child { margin-bottom: 0; }
    .t-progress-label { display: flex; justify-content: space-between; font-size: 0.875rem; font-weight: 600; color: #374151; margin-bottom: 0.25rem; }
    .t-progress-label .t-bold { color: #111827; font-size: 0.75rem; }
    .t-progress-bar-bg { width: 100%; background: rgba(191,219,254,0.5); border-radius: 9999px; height: 0.75rem; overflow: hidden; position: relative; z-index: 10; }
    .t-progress-bar-fill-blue { height: 100%; border-radius: 9999px; background: #3b82f6; }
    .t-progress-bar-fill-blue-light { height: 100%; border-radius: 9999px; background: #60a5fa; }
    
    .t-stats-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 0.75rem; margin-bottom: 1rem; position: relative; z-index: 10; }
    .t-stat-box { background: white; padding: 0.75rem; border-radius: 0.75rem; box-shadow: 0 1px 2px rgba(0,0,0,0.05); border: 1px solid rgba(255,237,213,0.5); text-align: center; display: flex; flex-direction: column; justify-content: center; }
    .t-stat-val { font-size: 1.5rem; font-weight: 900; line-height: 1; }
    .t-text-orange { color: #f97316; }
    .t-text-green { color: #10b981; }
    .t-text-yellow { color: #f59e0b; }
    .t-text-red { color: #ef4444; }
    .t-stat-lbl { font-size: 0.75rem; font-weight: 700; color: #6b7280; margin-top: 0.25rem; }
    
    .t-murojaah-bar-wrap { position: relative; z-index: 10; }
    .t-murojaah-lbl { display: flex; justify-content: space-between; font-size: 0.75rem; font-weight: 700; color: #4b5563; margin-bottom: 0.375rem; }
    .t-card-orange .t-progress-bar-bg { background: rgba(254,215,170,0.5); }
    .t-flex-bar { display: flex; }
    .t-bar-green { height: 100%; background: #10b981; border-right: 1px solid rgba(255,255,255,0.3); }
    .t-bar-yellow { height: 100%; background: #f59e0b; border-right: 1px solid rgba(255,255,255,0.3); }
    .t-bar-red { height: 100%; background: #ef4444; }
    .t-murojaah-legend { display: flex; justify-content: space-between; margin-top: 0.5rem; font-size: 0.625rem; font-weight: 700; }
    
    /* Daily View */
    .t-daily-view { display: flex; flex-direction: column; padding: 1.5rem; background: white; width: 100%; box-sizing: border-box; }
    @media (min-width: 768px) { .t-daily-view { padding: 2rem; border-bottom-left-radius: 1.5rem; border-bottom-right-radius: 1.5rem; } }
    .t-btn-back-outline { font-size: 0.875rem; background: white; border: 2px solid #e5e7eb; color: #374151; padding: 0.375rem 1rem; border-radius: 9999px; font-weight: 700; cursor: pointer; transition: 0.2s; }
    .t-btn-back-outline:hover { background: #f3f4f6; }
    
    .t-timeline { position: relative; border-left: 2px solid #d1fae5; margin-left: 0.75rem; padding-bottom: 1rem; }
    @media (min-width: 768px) { .t-timeline { margin-left: 1rem; } }
    .t-timeline-item { margin-bottom: 1.5rem; margin-left: 1.5rem; position: relative; }
    .t-timeline-dot { position: absolute; left: -2.1875rem; top: 0.375rem; width: 1.25rem; height: 1.25rem; border-radius: 9999px; border: 4px solid white; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
    .t-dot-green { background: #10b981; }
    .t-dot-orange { background: #fb923c; }
    .t-timeline-date { font-size: 0.875rem; font-weight: 700; color: #1f2937; margin-bottom: 0.5rem; }
    .t-timeline-box { padding: 0.75rem; border-radius: 0.75rem; margin-bottom: 0.5rem; box-shadow: 0 1px 2px rgba(0,0,0,0.05); }
    .t-box-blue { background: #eff6ff; border: 1px solid #dbeafe; }
    .t-box-orange { background: #fff7ed; border: 1px solid #ffedd5; }
    .t-badge-sm-blue { font-size: 0.625rem; font-weight: 700; color: #1d4ed8; background: #dbeafe; padding: 0.125rem 0.5rem; border-radius: 0.25rem; margin-right: 0.5rem; text-transform: uppercase; letter-spacing: 0.025em; }
    .t-badge-sm-orange { font-size: 0.625rem; font-weight: 700; color: #c2410c; background: #ffedd5; padding: 0.125rem 0.5rem; border-radius: 0.25rem; margin-right: 0.5rem; text-transform: uppercase; letter-spacing: 0.025em; }
    .t-box-text { font-size: 0.875rem; color: #374151; font-weight: 500; }
    
    /* Animations */
    @keyframes fillProgress { from { width: 0; } }
    :global(.animate-progress) { animation: fillProgress 1.5s ease-out forwards; }
    :global(.animate-circle) { transition: stroke-dashoffset 1.5s ease-out; }
</style>
