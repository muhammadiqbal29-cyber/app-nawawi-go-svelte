<script lang="ts">
    import { onMount } from 'svelte';
    import Icon from './Icon.svelte';

    // Props using Svelte 5 runic-style/attributes structure or Svelte 4 exports
    export let show: boolean = false;
    export let message: string = '';
    export let type: 'success' | 'error' | 'info' = 'success';
    export let duration: number = 3000;
    export let onClose: () => void = () => {};

    let timeoutId: any;

    $: if (show) {
        if (timeoutId) clearTimeout(timeoutId);
        timeoutId = setTimeout(() => {
            show = false;
            onClose();
        }, duration);
    }
</script>

{#if show}
    <div class="toast-notification toast-{type}">
        <div class="toast-icon">
            {#if type === 'success'}
                <Icon name="check" size="1.25rem" />
            {:else if type === 'error'}
                <Icon name="calendar-xmark" size="1.25rem" />
            {:else}
                <Icon name="clock" size="1.25rem" />
            {/if}
        </div>
        <div class="toast-message">{message}</div>
    </div>
{/if}

<style>
    .toast-notification {
        position: fixed;
        bottom: 5.5rem; /* Just above bottom nav if present, or near bottom */
        left: 50%;
        transform: translateX(-50%);
        background: #1f2937;
        color: white;
        padding: 0.875rem 1.25rem;
        border-radius: 1rem;
        box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
        display: flex;
        align-items: center;
        gap: 0.75rem;
        z-index: 1000;
        max-width: 90%;
        width: max-content;
        animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    .toast-success {
        background-color: #065f46; /* Deep emerald */
        border-color: #059669;
    }

    .toast-error {
        background-color: #991b1b; /* Deep red */
        border-color: #dc2626;
    }

    .toast-info {
        background-color: #1e3a8a; /* Deep blue */
        border-color: #2563eb;
    }

    .toast-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }

    .toast-message {
        font-size: 0.875rem;
        font-weight: 600;
        letter-spacing: 0.01em;
        line-height: 1.25;
    }

    @keyframes slideUp {
        from {
            opacity: 0;
            transform: translate(-50%, 1rem);
        }
        to {
            opacity: 1;
            transform: translate(-50%, 0);
        }
    }
</style>
