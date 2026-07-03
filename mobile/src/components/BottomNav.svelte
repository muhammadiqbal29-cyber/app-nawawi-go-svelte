<script lang="ts">
    import { navigate, routerState } from "../store/nav.svelte";
    import Icon from "./Icon.svelte";

    export let activeMenu = "home";
</script>

<nav class="bottom-nav">
    <button
    <button 
        class="nav-item {activeMenu === 'home' ? 'active' : ''}" 
        onclick={() => navigate('menu_utama')}
    >
        <Icon name="house" class="nav-icon" />
        <span class="nav-text">Home</span>
    </button>
    <button 
        class="nav-item {activeMenu === 'akademik' ? 'active' : ''}" 
        onclick={() => navigate('akademik')}
    >
        <Icon name="graduation-cap" class="nav-icon" />
        <span class="nav-text">Akademik</span>
    </button>
    <button 
        class="nav-item {activeMenu === 'izin' ? 'active' : ''}" 
        onclick={() => navigate('izin')}
    >
        <Icon name="calendar-plus" class="nav-icon" />
        <span class="nav-text">Izin</span>
    </button>
    <button 
        class="nav-item {activeMenu === 'riwayat' ? 'active' : ''}" 
        onclick={() => navigate('riwayat')}
    >
        <Icon name="clock-rotate-left" class="nav-icon" />
        <span class="nav-text">Riwayat</span>
    </button>
    <button 
        class="nav-item {activeMenu === 'profil' ? 'active' : ''}" 
        onclick={() => navigate('profil')}
    >
        <Icon name="user" class="nav-icon" />
        <span class="nav-text">Profil</span>
    </button>
</nav>

<style>
    .bottom-nav {
        background-color: #ffffff;
        border-top: 1px solid #e5e7eb;
        position: fixed;
        bottom: 0;
        width: 100%;
        max-width: 900px;
        display: flex;
        justify-content: space-around;
        align-items: center;
        z-index: 50;
        padding: 0 8px;
        padding-bottom: env(safe-area-inset-bottom);
        box-shadow: 0 -5px 15px rgba(0, 0, 0, 0.05);
        left: 50%;
        transform: translateX(-50%);
    }

    .nav-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        width: 100%;
        padding: 12px 0;
        color: #9ca3af;
        background: transparent;
        border: none;
        cursor: pointer;
        transition: color 0.3s;
    }

    .nav-item:hover,
    .nav-item.active {
        color: #059669;
    }

    :global(.nav-icon) {
        width: 1.25rem;
        height: 1.25rem;
        margin-bottom: 4px;
    }

    .nav-text {
        font-size: 10px;
        font-weight: 700;
        font-family: "Inter", sans-serif;
    }
</style>
