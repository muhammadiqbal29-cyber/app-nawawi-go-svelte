<script lang="ts">
    import { navigate } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import { onMount } from 'svelte';
    import { getCurrentGuru } from '../services/endpoints';

    let guruData = $state<any>(null);
    let aksesFitur = $state<string[]>([]);

    onMount(async () => {
        try {
            guruData = await getCurrentGuru();
            if (guruData && guruData.akses_fitur) {
                aksesFitur = JSON.parse(guruData.akses_fitur);
            }
        } catch(e) {
            console.error("Gagal load guru profil", e);
        }
    });
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">

        <!-- Header Clean -->
        <div class="page-header">
            <button class="back-btn" onclick={() => navigate('akademik')}>
                <Icon name="arrow-left" /> Kembali
            </button>
            <h1 class="page-title">Perkembangan</h1>
            <p class="page-subtitle">Pilih kategori laporan siswa</p>
        </div>

        <!-- Menu List -->
        <div class="scrollable-content">
            <div class="content-padding">
                <div class="menu-list">

                    <!-- Akademik Umum Removed -->

                    <!-- Tahsin & Tahfiz -->
                    {#if aksesFitur.includes('Tahfiz') || aksesFitur.length === 0}
                    <button class="menu-card bg-cyan" onclick={() => navigate('tahfiz')}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="quran" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Tahsin & Tahfiz</h3>
                            <p class="card-desc">Progress setoran & hafalan</p>
                        </div>
                    </button>
                    {/if}

                    <!-- Keseharian -->
                    {#if aksesFitur.includes('Keseharian') || aksesFitur.includes('Asrama') || aksesFitur.length === 0}
                    <button class="menu-card bg-purple" onclick={() => navigate('daftar_kelas_keseharian')}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="clipboard-check" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Keseharian</h3>
                            <p class="card-desc">Log aktivitas & pelanggaran</p>
                        </div>
                    </button>
                    {/if}

                </div>
            </div>
        </div>

    </div>
</div>

<style>
    .mobile-app-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        width: 100vw;
        background-color: #f3f4f6;
        font-family: 'Inter', sans-serif;
    }

    .app-container {
        background-color: #f9fafb;
        width: 100%;
        min-height: 100vh;
        position: relative;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    @media (min-width: 768px) {
        .mobile-app-wrapper { padding: 3rem 1.5rem; }
        .app-container {
            max-width: 900px;
            min-height: 0;
            max-height: 90vh;
            border-radius: 1.5rem;
            border: 1px solid #e5e7eb;
            box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1);
        }
    }

    /* Header Clean */
    .page-header {
        background-color: #f9fafb;
        padding: 1.5rem;
        position: relative;
        z-index: 10;
    }
    @media (min-width: 768px) { .page-header { padding: 2rem; } }

    .back-btn {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        color: #6b7280;
        font-weight: 600;
        font-size: 0.875rem;
        background: transparent;
        border: none;
        cursor: pointer;
        margin-bottom: 1rem;
        transition: color 0.2s;
    }
    .back-btn:hover { color: #10b981; }

    .page-title {
        font-size: 1.875rem;
        font-weight: 900;
        color: #1f2937;
        letter-spacing: 0.025em;
        margin: 0;
    }
    .page-subtitle {
        font-size: 0.875rem;
        font-weight: 600;
        color: #6b7280;
        margin-top: 0.25rem;
        margin-bottom: 0;
    }

    /* Content */
    .scrollable-content {
        flex-grow: 1;
        overflow-y: auto;
        -webkit-overflow-scrolling: touch;
        padding-bottom: 2rem;
    }
    .content-padding { padding: 0 1.5rem 1.5rem; }
    @media (min-width: 768px) { .content-padding { padding: 0 2rem 2rem; } }

    .menu-list { display: flex; flex-direction: column; gap: 1rem; }
    @media (min-width: 768px) { .menu-list { gap: 1.5rem; } }

    .menu-card {
        display: flex;
        align-items: center;
        width: 100%;
        padding: 1.25rem;
        border-radius: 2rem;
        border: none;
        cursor: pointer;
        position: relative;
        overflow: hidden;
        transition: transform 0.2s;
        color: white;
        text-align: left;
    }
    @media (min-width: 768px) { .menu-card { padding: 1.5rem; } }
    .menu-card:active { transform: scale(0.96); }

    .card-bubble {
        position: absolute;
        right: -2rem;
        top: -2rem;
        width: 8rem;
        height: 8rem;
        background: rgba(255,255,255,0.2);
        border-radius: 9999px;
        opacity: 0.5;
        transition: transform 0.5s;
    }
    .menu-card:hover .card-bubble { transform: scale(1.5); }

    .card-icon {
        width: 4rem;
        height: 4rem;
        background: rgba(255,255,255,0.2);
        border: 1px solid rgba(255,255,255,0.2);
        border-radius: 1rem;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.875rem;
        flex-shrink: 0;
        position: relative;
        z-index: 10;
    }
    .card-content { margin-left: 1.25rem; position: relative; z-index: 10; }
    .card-title { font-weight: 700; font-size: 1.25rem; margin: 0; }
    .card-desc { font-size: 0.875rem; color: rgba(255,255,255,0.9); font-weight: 500; margin: 0.25rem 0 0; }

    .bg-purple { background-color: #a855f7; box-shadow: 0 10px 25px -5px rgba(168,85,247,0.3); }
    .bg-cyan   { background-color: #06b6d4; box-shadow: 0 10px 25px -5px rgba(6,182,212,0.3); }
</style>
