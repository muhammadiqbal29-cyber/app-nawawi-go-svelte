<script lang="ts">
    import { navigate } from '../store/nav.svelte';
    import { onMount } from 'svelte';
    import BottomNav from '../components/BottomNav.svelte';
    import Icon from '../components/Icon.svelte';
    import { getCurrentGuru } from '../services/endpoints';

    let guruData = $state<any>(null);

    function handleLogout() {
        localStorage.removeItem('token');
        localStorage.removeItem('role');
        navigate('login');
    }

    onMount(async () => {
        try {
            guruData = await getCurrentGuru();
        } catch (error) {
            console.error("Failed to fetch profil:", error);
        }
    });
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        <div class="scrollable-content">
            <div class="content-padding">

                <!-- Profil Header -->
                <div class="profil-header">
                    {#if guruData}
                        <div class="avatar">{guruData.nama_lengkap?.substring(0,2).toUpperCase() || 'GU'}</div>
                        <h1 class="profil-name">{guruData.nama_lengkap}</h1>
                        <p class="profil-nip">{guruData.email}</p>
                        <div class="profil-role-badge">Tahun Masuk: {guruData.tahun_masuk || '-'}</div>
                    {:else}
                        <div class="avatar">...</div>
                        <h1 class="profil-name">Memuat...</h1>
                    {/if}
                </div>

                <!-- Menu List -->
                <div class="menu-list">

                    <!-- Ganti Password -->
                    <button class="menu-card bg-rose" onclick={() => navigate('ganti_password')}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="lock" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Ganti Password</h3>
                            <p class="card-desc">Perbarui kata sandi akun</p>
                        </div>
                    </button>

                    <!-- Logout -->
                    <button class="menu-card bg-red mt-large" onclick={handleLogout}>
                        <div class="card-bubble"></div>
                        <div class="card-icon">
                            <Icon name="arrow-right-from-bracket" />
                        </div>
                        <div class="card-content">
                            <h3 class="card-title">Logout</h3>
                            <p class="card-desc">Keluar dari aplikasi</p>
                        </div>
                    </button>

                </div>
            </div>
        </div>

        <BottomNav activeMenu="profil" />
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

    /* Profil Header */
    .profil-header {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-bottom: 2rem;
    }
    .avatar {
        width: 6rem;
        height: 6rem;
        background-color: #d1fae5;
        color: #059669;
        border-radius: 9999px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.875rem;
        font-weight: 900;
        border: 4px solid white;
        box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1);
        margin-bottom: 1rem;
    }
    .profil-name {
        font-size: 1.5rem;
        font-weight: 900;
        color: #1f2937;
        letter-spacing: 0.025em;
        margin: 0;
    }
    .profil-nip {
        font-size: 0.875rem;
        font-weight: 600;
        color: #6b7280;
        margin: 0.25rem 0 0;
    }
    .profil-role-badge {
        margin-top: 0.75rem;
        background-color: #ecfdf5;
        color: #065f46;
        padding: 0.25rem 1rem;
        border-radius: 9999px;
        font-size: 0.75rem;
        font-weight: 700;
        border: 1px solid #a7f3d0;
    }

    /* Menu Cards */
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
    .mt-large { margin-top: 2rem; }

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

    .bg-rose { background-color: #f43f5e; box-shadow: 0 10px 25px -5px rgba(244,63,94,0.3); }
    .bg-red  { background-color: #dc2626; box-shadow: 0 10px 25px -5px rgba(220,38,38,0.3); }
</style>
