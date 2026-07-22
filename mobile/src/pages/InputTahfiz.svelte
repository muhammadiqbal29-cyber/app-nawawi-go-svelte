<script lang="ts">
    import { navigate, routerState } from '../store/nav.svelte';
    import { onMount } from 'svelte';
    import Icon from '../components/Icon.svelte';
    import Toast from '../components/Toast.svelte';
    import { postSetoranTahfiz, getMateriTahfizAktif } from '../services/endpoints';
    import { guruState } from '../store/guru.svelte';

    let showToast = $state(false);
    let toastMessage = $state('');
    let toastType = $state<'success' | 'error' | 'info'>('success');
    let isLoading = $state(false);

    let idSiswa = $derived(routerState.navParams?.id_siswa);
    let idKelompok = $derived(routerState.navParams?.id_kelompok);
    let namaSiswa = $derived(routerState.navParams?.nama_siswa || 'Siswa');

    let jenisSetoran = $state<'Ziyadah' | 'Murojaah'>('Ziyadah');
    let catatan = $state('');
    let surah = $state('');
    let juz = $state<number>(30);
    
    let materiTahfizList = $state<any[]>([]);
    let selectedMateriId = $state<number | null>(null);

    onMount(async () => {
        try {
            materiTahfizList = await getMateriTahfizAktif(guruState.activeSekolahId || 1);
        } catch (e) {
            console.error("Gagal load materi", e);
        }
    });

    $effect(() => {
        if (selectedMateriId) {
            const m = materiTahfizList.find(x => x.id === selectedMateriId);
            if (m) {
                surah = m.nama_materi;
                juz = m.juz;
            }
        }
    });

    // Scores per indicator (1-5)
    const indicators = [
        { id: 'hafal',   title: 'Daya Ingat (Hafalan)',     icon: '🧠', options: ['Mengulang Total','Banyak Bimbingan','Sering Lupa','Hafal Baik','Mutqin (Sangat Kuat)'] },
        { id: 'lancar',  title: 'Kelancaran (Tempo/Aliran)', icon: '🌊', options: ['Terbata-bata','Sering Tersendat','Cukup Lancar','Lancar','Sangat Mengalir'] },
        { id: 'mad',     title: 'Mad (Panjang Pendek)',      icon: '📏', options: ['Sangat Berantakan','Banyak Tertukar','Cukup Tepat','Hampir Sempurna','Sempurna Konsisten'] },
        { id: 'tajwid',  title: 'Tajwid (Hukum Bacaan)',    icon: '📖', options: ['Belum Paham','Sering Keliru','Cukup Tepat','Baik','Sempurna'] },
        { id: 'makhroj', title: 'Makharijul Huruf',          icon: '🗣️', options: ['Kurang Jelas','Banyak Tertukar','Cukup Jelas','Fasih','Sangat Fasih'] },
    ];

    let scores = $state<Record<string, number>>({
        hafal: 0, lancar: 0, mad: 0, tajwid: 0, makhroj: 0
    });

    const scoreColors = ['','text-red','text-orange','text-yellow','text-blue','text-emerald'];
    const scoreCardColors: Array<{ bg: string; border: string; text: string }> = [
        { bg: '', border: '', text: '' },
        { bg: '#fef2f2', border: '#fecaca', text: '#dc2626' },
        { bg: '#fff7ed', border: '#fed7aa', text: '#ea580c' },
        { bg: '#fefce8', border: '#fde68a', text: '#ca8a04' },
        { bg: '#eff6ff', border: '#bfdbfe', text: '#2563eb' },
        { bg: '#ecfdf5', border: '#a7f3d0', text: '#059669' },
    ];

    let finalScore = $derived(
        Object.values(scores).reduce((a, b) => a + b, 0) * 4
    );

    let predikat = $derived(() => {
        if (finalScore >= 85) return 'Mumtaz';
        if (finalScore >= 70) return 'Jayyid';
        if (finalScore >= 55) return 'Maqbul';
        return 'Rasib';
    });

    async function triggerSave() {
        if (!idSiswa || !idKelompok) {
            toastMessage = 'Data siswa tidak valid';
            toastType = 'error';
            showToast = true;
            return;
        }

        isLoading = true;
        try {
            await postSetoranTahfiz({
                id_siswa: parseInt(idSiswa),
                id_kelompok: parseInt(idKelompok),
                juz: juz,
                surah: surah || '-',
                jenis_setoran: jenisSetoran,
                skor_hafalan: scores.hafal,
                skor_kelancaran: scores.lancar,
                skor_mad: scores.mad,
                skor_tajwid: scores.tajwid,
                skor_makharijul_huruf: scores.makhroj,
                total_poin: finalScore,
                predikat: predikat(),
                catatan: catatan
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
            <h1 class="header-title">Input Penilaian</h1>
            <div class="header-badge">{namaSiswa}</div>
        </div>

        <!-- Content -->
        <div class="scrollable-content">
            <div class="content-padding">

                <!-- Jenis Setoran Toggle -->
                <div class="form-group">
                    <fieldset class="form-fieldset">
                        <legend class="form-label">Jenis Setoran</legend>
                        <div class="toggle-wrap">
                        <div class="toggle-bg" style="transform: translateX({jenisSetoran === 'Murojaah' ? '100%' : '0'})"></div>
                        <button class="toggle-btn {jenisSetoran === 'Ziyadah' ? 'active' : ''}" onclick={() => jenisSetoran = 'Ziyadah'}>Ziyadah</button>
                        <button class="toggle-btn {jenisSetoran === 'Murojaah' ? 'active' : ''}" onclick={() => jenisSetoran = 'Murojaah'}>Murojaah</button>
                        </div>
                    </fieldset>
                </div>

                <!-- Form Ziyadah -->
                {#if jenisSetoran === 'Ziyadah'}
                <div class="form-card">
                    <div class="form-group">
                        <label class="form-label-sm" for="materi-select">Pilih Surah / Materi</label>
                        <select id="materi-select" class="form-select" bind:value={selectedMateriId}>
                            <option value={null}>-- Pilih Materi --</option>
                            {#each materiTahfizList as m}
                                <option value={m.id}>{m.nama_materi} (Juz {m.juz})</option>
                            {/each}
                        </select>
                    </div>
                    <div class="ayat-row">
                        <div class="flex-1">
                            <label class="form-label-sm" for="ayat-mulai">Ayat Mulai</label>
                            <input id="ayat-mulai" type="number" value="1" class="form-input text-center" />
                        </div>
                        <div class="ayat-dash">-</div>
                        <div class="flex-1">
                            <label class="form-label-sm" for="ayat-sampai">Ayat Sampai</label>
                            <input id="ayat-sampai" type="number" value="5" class="form-input text-center" />
                        </div>
                    </div>
                </div>
                {:else}
                <!-- Form Murojaah -->
                <div class="form-card">
                    <div class="form-group">
                        <fieldset class="form-fieldset">
                            <legend class="form-label-sm">Mode Murojaah</legend>
                            <div class="mode-grid">
                            <button class="mode-btn mode-active">Per Juz</button>
                            <button class="mode-btn">Per Surah</button>
                            </div>
                        </fieldset>
                    </div>
                    <div class="form-group">
                        <label class="form-label-sm" for="juz-select">Pilih Juz</label>
                        <input id="juz-select" type="number" class="form-input" placeholder="Ketik angka juz..." bind:value={juz} />
                    </div>
                    <div>
                        <label class="form-label-sm" for="catatan-rentang">Catatan Rentang (Opsional)</label>
                        <input id="catatan-rentang" type="text" placeholder="Misal: An-Naba s/d At-Takwir" class="form-input" bind:value={surah} />
                    </div>
                </div>
                {/if}

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
                <p class="nilai-label">Nilai Akhir</p>
                <div class="nilai-display">
                    <span class="nilai-number">{finalScore}</span>
                    <span class="nilai-max">/ 100</span>
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
    .mobile-app-wrapper {
        display: flex; align-items: center; justify-content: center;
        min-height: 100vh; width: 100vw;
        background-color: #f3f4f6; font-family: 'Inter', sans-serif;
    }
    .app-container {
        background-color: #f9fafb; width: 100%; min-height: 100vh;
        position: relative; display: flex; flex-direction: column; overflow: hidden;
    }
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
    .header-badge { display: inline-block; background: rgba(255,255,255,0.2); padding: 0.375rem 1rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 700; }

    /* Content */
    .scrollable-content { flex-grow: 1; overflow-y: auto; -webkit-overflow-scrolling: touch; padding-bottom: 6rem; }
    .content-padding { padding: 1.5rem; }

    /* Form */
    .form-group { margin-bottom: 1.5rem; }
    .form-label { font-size: 0.75rem; font-weight: 700; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em; display: block; margin-bottom: 0.5rem; }
    .form-label-sm { font-size: 0.75rem; font-weight: 700; color: #6b7280; display: block; margin-bottom: 0.25rem; }
    .form-fieldset { border: 0; margin: 0; padding: 0; }

    .toggle-wrap { display: flex; background-color: #e5e7eb; padding: 0.25rem; border-radius: 0.75rem; position: relative; }
    .toggle-bg { position: absolute; top: 0.25rem; left: 0.25rem; bottom: 0.25rem; width: calc(50% - 0.25rem); background: white; border-radius: 0.5rem; box-shadow: 0 1px 2px rgba(0,0,0,0.1); transition: transform 0.3s ease; }
    .toggle-btn { flex: 1; padding: 0.5rem; font-size: 0.875rem; font-weight: 700; border-radius: 0.5rem; border: none; cursor: pointer; background: transparent; color: #6b7280; position: relative; z-index: 10; transition: color 0.3s; }
    .toggle-btn.active { color: #ec4899; }

    .form-card { background: white; padding: 1.25rem; border-radius: 1rem; border: 1px solid #f3f4f6; box-shadow: 0 1px 2px rgba(0,0,0,0.05); margin-bottom: 1.5rem; }
    .form-select, .form-input {
        width: 100%; background-color: #f9fafb; border: 1px solid #e5e7eb; color: #1f2937;
        font-size: 0.875rem; border-radius: 0.5rem; padding: 0.625rem; font-weight: 600;
        outline: none; font-family: inherit; box-sizing: border-box;
    }
    .form-input::-webkit-inner-spin-button, .form-input::-webkit-outer-spin-button { -webkit-appearance: none; }
    .ayat-row { display: flex; gap: 1rem; align-items: flex-end; }
    .ayat-dash { color: #9ca3af; font-weight: 700; padding-bottom: 0.625rem; }
    .flex-1 { flex: 1; }
    .mode-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
    .mode-btn { padding: 0.5rem; border-radius: 0.5rem; font-size: 0.625rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; border: 1px solid #e5e7eb; background-color: #f9fafb; color: #6b7280; cursor: pointer; }
    .mode-active { background-color: #fdf2f8; color: #be185d; border-color: #fbcfe8; }
    .form-textarea { width: 100%; background: white; border: 1px solid #e5e7eb; border-radius: 1rem; padding: 1rem; outline: none; resize: none; font-family: inherit; font-size: 0.875rem; box-sizing: border-box; }
    .mt-lg { margin-top: 1.5rem; }

    /* Indicators */
    .indicator-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1rem; }
    .indicator-title { font-size: 1.125rem; font-weight: 900; color: #1f2937; margin: 0; }
    .skala-badge { font-size: 0.75rem; font-weight: 700; background: #fdf2f8; color: #be185d; padding: 0.25rem 0.5rem; border-radius: 0.5rem; }
    .indicators-list { display: flex; flex-direction: column; gap: 1rem; }
    .indicator-card { background: white; padding: 1rem; border-radius: 1rem; border: 1px solid #f3f4f6; box-shadow: 0 1px 2px rgba(0,0,0,0.05); }
    .indicator-head { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.75rem; }
    .indicator-icon { font-size: 1.25rem; }
    .indicator-name { font-weight: 700; color: #1f2937; font-size: 0.875rem; margin: 0; }
    .score-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 0.5rem; }
    @media (min-width: 768px) { .score-grid { grid-template-columns: repeat(5, 1fr); } }
    .score-btn {
        padding: 0.5rem 0.25rem; border-radius: 0.75rem; font-size: 0.625rem; font-weight: 700;
        border: 1px solid #e5e7eb; background-color: #f9fafb; color: #6b7280;
        cursor: pointer; transition: all 0.2s; text-align: center;
    }
    .score-btn.selected { transform: scale(1.05); }
    .score-num { display: block; font-size: 1rem; font-weight: 900; }
    .score-label { display: block; font-size: 0.6rem; margin-top: 0.125rem; line-height: 1.2; }

    /* Bottom Bar */
    .bottom-bar { position: absolute; bottom: 0; left: 0; width: 100%; background: white; border-top: 1px solid #f3f4f6; box-shadow: 0 -10px 20px rgba(0,0,0,0.05); padding: 1.25rem; z-index: 20; display: flex; align-items: center; justify-content: space-between; box-sizing: border-box; }
    .nilai-label { font-size: 0.75rem; font-weight: 700; color: #6b7280; margin: 0 0 0.25rem; }
    .nilai-display { display: flex; align-items: flex-end; gap: 0.25rem; }
    .nilai-number { font-size: 1.875rem; font-weight: 900; color: #ec4899; }
    .nilai-max { font-size: 0.875rem; font-weight: 700; color: #9ca3af; margin-bottom: 0.25rem; }
    .btn-simpan { background-color: #ec4899; color: white; font-weight: 700; padding: 0.75rem 1.5rem; border-radius: 1rem; border: none; cursor: pointer; display: flex; align-items: center; gap: 0.5rem; box-shadow: 0 10px 15px -3px rgba(236,72,153,0.3); transition: background-color 0.2s, transform 0.1s; font-family: inherit; font-size: 1rem; }
    .btn-simpan:hover { background-color: #db2777; }
    .btn-simpan:active { transform: scale(0.95); }

    /* Daily View */
    .daily-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
    .daily-title { font-size: 1.25rem; font-weight: 700; color: #1f2937; margin: 0; }
    .btn-kembali-daily { font-size: 0.875rem; font-weight: 700; background: white; border: 2px solid #e5e7eb; color: #374151; padding: 0.375rem 1rem; border-radius: 9999px; cursor: pointer; }
    .daily-timeline { border-left: 2px solid #d1fae5; margin-left: 0.75rem; padding-bottom: 1rem; }
    .timeline-item { margin-bottom: 1.5rem; margin-left: 1.5rem; position: relative; }
    .timeline-dot-wrap { position: absolute; left: -2.25rem; top: 0.375rem; }
    .dot { width: 1.25rem; height: 1.25rem; border-radius: 9999px; border: 4px solid white; box-shadow: 0 1px 2px rgba(0,0,0,0.1); }
    .dot-green { background-color: #10b981; }
    .dot-orange { background-color: #f97316; }
    .timeline-date { font-size: 0.875rem; font-weight: 700; color: #1f2937; margin-bottom: 0.5rem; }
    .timeline-entry { padding: 0.75rem; border-radius: 0.75rem; font-size: 0.875rem; margin-bottom: 0.5rem; }
    .entry-blue { background: #eff6ff; border: 1px solid #dbeafe; }
    .entry-orange { background: #fff7ed; border: 1px solid #fed7aa; }
    .entry-badge { font-size: 0.625rem; font-weight: 700; padding: 0.125rem 0.5rem; border-radius: 0.25rem; margin-right: 0.5rem; text-transform: uppercase; letter-spacing: 0.05em; }
    .badge-blue { background: #dbeafe; color: #1d4ed8; }
    .badge-orange { background: #fed7aa; color: #c2410c; }
    .entry-text { color: #374151; font-weight: 500; }
    .entry-status { font-weight: 700; }
</style>
