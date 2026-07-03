<script lang="ts">
    import { navigate } from '../store/nav.svelte';
    import BottomNav from '../components/BottomNav.svelte';
    import Icon from '../components/Icon.svelte';
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        <div class="scrollable-content">
            <div class="content-padding">

                <!-- Judul -->
                <div class="page-header">
                    <h1 class="page-title">Akademik</h1>
                    <p class="page-subtitle">Kelola nilai dan perkembangan siswa</p>
                </div>

                <!-- Menu List -->
                <div class="menu-list">

                    <!-- Perkembangan Siswa -->
                    <button class="menu-card bg-purple" onclick={() => navigate('menu_perkembangan')}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="chart-line" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Perkembangan Siswa</h3>
                            <p class="card-desc">Pantau rapor & nilai siswa</p>
                        </div>
                    </button>

                                        <!-- Input Nilai Tahfiz -->
                    <button class="menu-card bg-pink" onclick={() => navigate('daftar_santri_tahfiz')}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="pen-to-square" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Input Nilai Tahfiz</h3>
                            <p class="card-desc">Input nilai setoran hafalan harian</p>
                        </div>
                    </button>

                    <!-- Input Keseharian -->
                    <button class="menu-card bg-emerald" onclick={() => navigate('daftar_kelas_keseharian')}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="clipboard-check" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Ceklis Keseharian</h3>
                            <p class="card-desc">Input kehadiran/ceklis indikator harian</p>
                        </div>
                    </button>

                </div>
            </div>
        </div>

        <BottomNav activeMenu="akademik" />
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

    .scrollable-content {
        flex-grow: 1;
        overflow-y: auto;
        -webkit-overflow-scrolling: touch;
        padding-bottom: 7rem;
        padding-top: 2rem;
    }

    .content-padding { padding: 1.5rem; }
    @media (min-width: 768px) { .content-padding { padding: 2rem; } }

    .page-header { margin-bottom: 2rem; padding-left: 0.25rem; }
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

    .menu-list {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
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
        transition: transform 0.2s, box-shadow 0.2s;
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

    .card-content {
        margin-left: 1.25rem;
        position: relative;
        z-index: 10;
    }
    .card-title { font-weight: 700; font-size: 1.25rem; margin: 0; }
    .card-desc { font-size: 0.875rem; color: rgba(255,255,255,0.9); font-weight: 500; margin: 0.25rem 0 0; }

    .bg-blue   { background-color: #3b82f6; box-shadow: 0 10px 25px -5px rgba(59,130,246,0.3); }
    .bg-purple { background-color: #a855f7; box-shadow: 0 10px 25px -5px rgba(168,85,247,0.3); }
    .bg-emerald{ background-color: #10b981; box-shadow: 0 10px 25px -5px rgba(16,185,129,0.3); }
    .bg-pink   { background-color: #ec4899; box-shadow: 0 10px 25px -5px rgba(236,72,153,0.3); }
</style>
