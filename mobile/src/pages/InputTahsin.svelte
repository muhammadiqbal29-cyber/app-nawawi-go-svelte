<script lang="ts">
    import { navigate, routerState } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import Toast from '../components/Toast.svelte';
    import { postSetoranTahsin } from '../services/endpoints';

    let showToast = $state(false);
    let toastMessage = $state('');
    let toastType = $state<'success' | 'error' | 'info'>('success');
    let isLoading = $state(false);

    let idSiswa = $derived(routerState.navParams?.id_siswa);
    let idKelompok = $derived(routerState.navParams?.id_kelompok);
    let namaSiswa = $derived(routerState.navParams?.nama_siswa || 'Siswa');

    let bukuJilid = $state('Iqro 1');
    let halamanAtauSurah = $state('');
    let catatan = $state('');

    const bukuOptions = ['Iqro 1', 'Iqro 2', 'Iqro 3', 'Iqro 4', 'Iqro 5', 'Iqro 6', 'Al-Quran'];

    // Scores per indicator (1-5) for Tahsin
    const indicators = [
        { id: 'makhroj', title: 'Makharijul Huruf',          icon: '🗣️', options: ['Kurang Jelas','Banyak Tertukar','Cukup Jelas','Fasih','Sangat Fasih'] },
        { id: 'tajwid',  title: 'Tajwid (Hukum Bacaan)',    icon: '📖', options: ['Belum Paham','Sering Keliru','Cukup Tepat','Baik','Sempurna'] },
        { id: 'kelancaran',title: 'Kelancaran Membaca',      icon: '🌊', options: ['Terbata-bata','Sering Tersendat','Cukup Lancar','Lancar','Sangat Mengalir'] },
    ];

    let scores = $state<Record<string, number>>({
        makhroj: 0, tajwid: 0, kelancaran: 0
    });

    const scoreCardColors: Array<{ bg: string; border: string; text: string }> = [
        { bg: '', border: '', text: '' },
        { bg: '#fef2f2', border: '#fecaca', text: '#dc2626' },
        { bg: '#fff7ed', border: '#fed7aa', text: '#ea580c' },
        { bg: '#fefce8', border: '#fde68a', text: '#ca8a04' },
        { bg: '#eff6ff', border: '#bfdbfe', text: '#2563eb' },
        { bg: '#ecfdf5', border: '#a7f3d0', text: '#059669' },
    ];

    let finalScore = $derived(
        Object.values(scores).reduce((a, b) => a + b, 0)
    );
    let maxScore = indicators.length * 5;

    let predikat = $derived(() => {
        const pct = (finalScore / maxScore) * 100;
        if (pct >= 80) return 'Lancar';
        if (pct >= 60) return 'Kurang Lancar';
        return 'Ulang';
    });

    async function triggerSave() {
        if (!idSiswa || !idKelompok) {
            toastMessage = 'Data siswa tidak valid';
            toastType = 'error';
            showToast = true;
            return;
        }
        if (!halamanAtauSurah) {
            toastMessage = 'Halaman / Surah wajib diisi';
            toastType = 'error';
            showToast = true;
            return;
        }

        isLoading = true;
        try {
            await postSetoranTahsin({
                id_siswa: parseInt(idSiswa),
                id_kelompok: parseInt(idKelompok),
                buku_jilid: bukuJilid,
                halaman_atau_surah: halamanAtauSurah,
                predikat: predikat()
            });
            toastMessage = 'Berhasil disimpan!';
            toastType = 'success';
            showToast = true;
            
            setTimeout(() => {
                navigate('daftar_santri_tahfiz');
            }, 1500);
        } catch (e: any) {
            toastMessage = e.message || 'Gagal menyimpan setoran';
            toastType = 'error';
            showToast = true;
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">

        <!-- Header Pink -->
        <div class="page-header">
            <button class="back-btn" onclick={() => navigate('daftar_santri_tahfiz')}>
                <Icon name="arrow-left" /> Kembali
            </button>
            <h1 class="header-title">Input Tahsin</h1>
            <div class="header-badge">{namaSiswa}</div>
        </div>

        <!-- Content -->
        <div class="scrollable-content">
            <div class="content-padding">

                <div class="form-card">
                    <div class="form-group">
                        <label class="form-label-sm" for="buku-select">Buku / Jilid</label>
                        <select id="buku-select" class="form-select" bind:value={bukuJilid}>
                            {#each bukuOptions as opt}
                                <option value={opt}>{opt}</option>
                            {/each}
                        </select>
                    </div>
                    <div class="form-group">
                        <label class="form-label-sm" for="halaman-input">Halaman / Surah</label>
                        <input id="halaman-input" type="text" class="form-input" placeholder="Misal: Hal 12 atau Surah Al-Baqarah" bind:value={halamanAtauSurah} />
                    </div>
                </div>

                <!-- Indikator Penilaian -->
                <div class="indicator-header">
                    <h3 class="indicator-title">Indikator Penilaian</h3>
                    <span class="skala-badge">Skala 1 - 5</span>
                </div>

                <div class="indicators-list">
                    {#each indicators as ind}
                    <div class="indicator-card">
                        <div class="indicator-head">
                            <span class="indicator-icon">{ind.icon}</span>
                            <h4 class="indicator-name">{ind.title}</h4>
                        </div>
                        <div class="score-grid">
                            {#each ind.options as opt, i}
                            {@const val = i + 1}
                            {@const color = scoreCardColors[val]}
                            <button
                                class="score-btn {scores[ind.id] === val ? 'selected' : ''}"
                                style={scores[ind.id] === val ? `background-color: ${color.bg}; border-color: ${color.border}; color: ${color.text};` : ''}
                                onclick={() => scores[ind.id] = val}
                            >
                                <span class="score-num">{val}</span>
                                <span class="score-label">{opt}</span>
                            </button>
                            {/each}
                        </div>
                    </div>
                    {/each}
                </div>

                <!-- Catatan -->
                <div class="form-group mt-lg">
                    <label class="form-label-sm" for="catatan-guru">Catatan Guru (Opsional)</label>
                    <textarea id="catatan-guru" rows="3" placeholder="Tambahkan catatan khusus untuk santri..." class="form-textarea" bind:value={catatan}></textarea>
                </div>

            </div>
        </div>

        <!-- Bottom Bar -->
        <div class="bottom-bar">
            <div>
                <p class="nilai-label">Status Kelulusan</p>
                <div class="nilai-display">
                    <span class="nilai-status" style="color: {predikat() === 'Lancar' ? '#10b981' : predikat() === 'Kurang Lancar' ? '#f59e0b' : '#ef4444'}">
                        {predikat()}
                    </span>
                </div>
            </div>
            <button class="btn-simpan" onclick={triggerSave} disabled={isLoading}>
                {isLoading ? 'Menyimpan...' : 'Simpan Nilai'}
                <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                </svg>
            </button>
        </div>

        <Toast bind:show={showToast} message={toastMessage} type={toastType} />

    </div>
</div>

<style>
    .mobile-app-wrapper { display: flex; align-items: center; justify-content: center; min-height: 100vh; width: 100vw; background-color: #f3f4f6; font-family: 'Inter', sans-serif; }
    .app-container { background-color: #f9fafb; width: 100%; min-height: 100vh; position: relative; display: flex; flex-direction: column; overflow: hidden; }
    @media (min-width: 768px) {
        .mobile-app-wrapper { padding: 3rem 1.5rem; }
        .app-container { max-width: 900px; min-height: 0; max-height: 90vh; border-radius: 1.5rem; border: 1px solid #e5e7eb; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1); }
    }

    /* Header */
    .page-header { background-color: #ec4899; color: white; padding: 1.5rem; text-align: center; position: relative; z-index: 10; border-bottom: 4px solid #db2777; }
    @media (min-width: 768px) { .page-header { padding: 2rem; } }
    .back-btn { position: absolute; top: 1.5rem; left: 1.5rem; color: rgba(255,255,255,0.8); background: transparent; border: none; cursor: pointer; font-size: 0.875rem; font-weight: 600; display: flex; align-items: center; gap: 0.5rem; }
    .back-btn:hover { color: white; }
    .header-title { font-size: 1.5rem; font-weight: 900; margin: 2rem 0 0.75rem; }
    @media (min-width: 768px) { .header-title { margin-top: 0; } }
    .header-badge { display: inline-block; background: rgba(255,255,255,0.2); padding: 0.375rem 1rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 700; letter-spacing: 0.05em; }

    /* Content */
    .scrollable-content { flex-grow: 1; overflow-y: auto; -webkit-overflow-scrolling: touch; padding-bottom: 6rem; }
    .content-padding { padding: 1.5rem; max-width: 600px; margin: 0 auto; }

    .form-card { background: white; padding: 1.25rem; border-radius: 1rem; box-shadow: 0 1px 2px rgba(0,0,0,0.05); border: 1px solid #f3f4f6; margin-bottom: 1.5rem; }
    .form-group { margin-bottom: 1.25rem; }
    .form-group:last-child { margin-bottom: 0; }
    .form-label-sm { display: block; font-size: 0.75rem; font-weight: 700; color: #6b7280; margin-bottom: 0.5rem; }
    .form-input, .form-select, .form-textarea { width: 100%; box-sizing: border-box; padding: 0.75rem 1rem; border-radius: 0.75rem; border: 1px solid #d1d5db; background: #f9fafb; font-size: 0.875rem; color: #1f2937; transition: 0.2s; font-family: inherit; }
    .form-input:focus, .form-select:focus, .form-textarea:focus { outline: none; border-color: #ec4899; box-shadow: 0 0 0 3px rgba(236,72,153,0.1); background: white; }

    /* Indicators */
    .indicator-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; margin-top: 2rem; }
    .indicator-title { font-size: 1.125rem; font-weight: 900; color: #1f2937; margin: 0; }
    .skala-badge { font-size: 0.625rem; font-weight: 700; background: #e5e7eb; color: #4b5563; padding: 0.25rem 0.5rem; border-radius: 9999px; text-transform: uppercase; letter-spacing: 0.05em; }
    
    .indicators-list { display: flex; flex-direction: column; gap: 1rem; }
    .indicator-card { background: white; padding: 1rem; border-radius: 1rem; border: 1px solid #e5e7eb; }
    .indicator-head { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem; }
    .indicator-icon { font-size: 1.25rem; }
    .indicator-name { font-size: 0.875rem; font-weight: 700; color: #374151; margin: 0; }
    
    .score-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 0.375rem; }
    .score-btn { padding: 0.5rem 0.25rem; background: white; border: 1px solid #d1d5db; border-radius: 0.5rem; cursor: pointer; transition: 0.2s; display: flex; flex-direction: column; align-items: center; gap: 0.25rem; }
    .score-num { font-size: 1rem; font-weight: 900; color: #9ca3af; }
    .score-label { font-size: 0.5rem; font-weight: 700; color: #9ca3af; text-align: center; text-transform: uppercase; line-height: 1.2; }
    
    .score-btn:hover { border-color: #9ca3af; }
    
    .mt-lg { margin-top: 1.5rem; }

    /* Bottom Bar */
    .bottom-bar { position: absolute; bottom: 0; left: 0; width: 100%; background: white; border-top: 1px solid #f3f4f6; padding: 1rem 1.5rem; display: flex; justify-content: space-between; align-items: center; box-shadow: 0 -10px 20px rgba(0,0,0,0.05); box-sizing: border-box; }
    .nilai-label { font-size: 0.625rem; font-weight: 700; color: #6b7280; text-transform: uppercase; margin: 0 0 0.25rem; }
    .nilai-display { display: flex; align-items: baseline; gap: 0.25rem; }
    .nilai-status { font-size: 1.25rem; font-weight: 900; }
    
    .btn-simpan { background-color: #ec4899; color: white; font-weight: 700; padding: 0.75rem 1.5rem; border-radius: 1rem; border: none; cursor: pointer; display: flex; align-items: center; gap: 0.5rem; box-shadow: 0 10px 15px -3px rgba(236,72,153,0.3); font-size: 1rem; transition: 0.2s; font-family: inherit; }
    .btn-simpan:hover { background-color: #db2777; }
    .btn-simpan:disabled { opacity: 0.7; cursor: not-allowed; }
</style>
