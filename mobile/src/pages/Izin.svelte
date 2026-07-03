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
                    <h1 class="page-title">Pengajuan Izin</h1>
                    <p class="page-subtitle">Kelola izin dan cuti Anda</p>
                </div>

                <!-- Empty State -->
                <div class="empty-state">
                    <div class="empty-icon">
                        <Icon name="folder-open" />
                    </div>
                    <h3 class="empty-title">Belum Ada Pengajuan Izin</h3>
                    <p class="empty-desc">Anda belum pernah mengajukan izin atau cuti pada semester ini.</p>
                    <button class="btn-ajukan">
                        <Icon name="plus" /> Ajukan Izin Baru
                    </button>
                </div>

            </div>
        </div>

        <BottomNav activeMenu="izin" />
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

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 3rem 0;
        text-align: center;
    }

    .empty-icon {
        width: 6rem;
        height: 6rem;
        background-color: #f3f4f6;
        border-radius: 9999px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #9ca3af;
        font-size: 2.5rem;
        margin-bottom: 1rem;
    }

    .empty-title {
        font-weight: 700;
        color: #1f2937;
        font-size: 1.125rem;
        margin: 0;
    }

    .empty-desc {
        color: #6b7280;
        font-size: 0.875rem;
        margin-top: 0.5rem;
        max-width: 20rem;
    }

    .btn-ajukan {
        margin-top: 2rem;
        background-color: #10b981;
        color: white;
        font-weight: 700;
        padding: 0.75rem 2rem;
        border-radius: 1rem;
        border: none;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        box-shadow: 0 10px 15px -3px rgba(16, 185, 129, 0.3);
        transition: background-color 0.2s, transform 0.1s;
        font-size: 1rem;
    }
    .btn-ajukan:hover { background-color: #059669; }
    .btn-ajukan:active { transform: scale(0.95); }
</style>
