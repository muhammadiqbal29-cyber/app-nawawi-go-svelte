<script lang="ts">
    import { toastStore } from '../store/toast.svelte';
    import { fade, slide } from 'svelte/transition';
</script>

<div class="toast-container">
    {#each toastStore.messages as toast (toast.id)}
        <div class="toast {toast.type}" in:slide={{ duration: 300 }} out:fade={{ duration: 200 }}>
            <div class="toast-icon">
                {#if toast.type === 'success'}
                    <i class="fas fa-check-circle"></i>
                {:else if toast.type === 'error'}
                    <i class="fas fa-exclamation-circle"></i>
                {:else if toast.type === 'warning'}
                    <i class="fas fa-exclamation-triangle"></i>
                {:else}
                    <i class="fas fa-info-circle"></i>
                {/if}
            </div>
            <div class="toast-content">
                {toast.text}
            </div>
        </div>
    {/each}
</div>

<style>
    .toast-container {
        position: fixed;
        top: 20px;
        left: 50%;
        transform: translateX(-50%);
        z-index: 9999;
        display: flex;
        flex-direction: column;
        gap: 10px;
        pointer-events: none;
        width: 90%;
        max-width: 400px;
    }

    .toast {
        display: flex;
        align-items: center;
        padding: 12px 16px;
        border-radius: 12px;
        box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
        color: white;
        font-weight: 500;
        font-size: 14px;
        backdrop-filter: blur(10px);
        -webkit-backdrop-filter: blur(10px);
        pointer-events: auto;
    }

    .toast-icon {
        margin-right: 12px;
        font-size: 18px;
    }

    .toast-content {
        flex: 1;
        line-height: 1.4;
    }

    .toast.success {
        background: rgba(16, 185, 129, 0.9);
        border: 1px solid rgba(16, 185, 129, 1);
    }

    .toast.error {
        background: rgba(239, 68, 68, 0.9);
        border: 1px solid rgba(239, 68, 68, 1);
    }

    .toast.warning {
        background: rgba(245, 158, 11, 0.9);
        border: 1px solid rgba(245, 158, 11, 1);
    }

    .toast.info {
        background: rgba(59, 130, 246, 0.9);
        border: 1px solid rgba(59, 130, 246, 1);
    }
</style>
