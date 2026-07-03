<script lang="ts">
    import Icon from './Icon.svelte';
</script>

<div class="landscape-overlay">
    <div class="landscape-message">
        <Icon name="rotate-left" class="rotate-icon" />
        <h2>Mohon Putar Perangkat Anda</h2>
        <p>Aplikasi ini dirancang untuk digunakan dalam mode <strong>Landscape</strong> (Mendatar) pada layar Tablet/iPad.</p>
    </div>
</div>

<style>
    .landscape-overlay {
        display: none;
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        background-color: #1f2937;
        color: white;
        z-index: 99999;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        text-align: center;
        padding: 2rem;
    }


    .landscape-message h2 {
        font-family: 'Inter', sans-serif;
        font-size: 1.5rem;
        font-weight: 700;
        margin-bottom: 0.5rem;
    }

    .landscape-message p {
        font-family: 'Inter', sans-serif;
        font-size: 1rem;
        color: #d1d5db;
        max-width: 400px;
        line-height: 1.5;
    }

    @keyframes rotate-anim {
        0% { transform: rotate(0deg); }
        50% { transform: rotate(-90deg); }
        100% { transform: rotate(-90deg); }
    }

    /* Tampilkan overlay HANYA pada tablet/iPad jika dalam mode portrait */
    /* Kita asumsikan perangkat dengan lebar layar >= 768px (Tablet) dan orientation portrait */
    @media screen and (min-width: 768px) and (orientation: portrait) {
        .landscape-overlay {
            display: flex;
        }
    }
</style>
