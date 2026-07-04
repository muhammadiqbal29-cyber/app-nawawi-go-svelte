<script lang="ts">
    import { modalStore } from '../store/modal.svelte';
    import { fade, scale } from 'svelte/transition';
</script>

{#if modalStore.isOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-overlay" transition:fade={{ duration: 200 }} onclick={modalStore.handleCancel}>
        <div class="modal-content" transition:scale={{ duration: 200, start: 0.95 }} onclick={(e) => e.stopPropagation()}>
            <div class="modal-header">
                <h3>{modalStore.title}</h3>
            </div>
            <div class="modal-body">
                <p>{modalStore.message}</p>
            </div>
            <div class="modal-footer">
                <button class="btn-cancel" onclick={modalStore.handleCancel}>Batal</button>
                <button class="btn-confirm" onclick={modalStore.handleConfirm}>Ya, Lanjutkan</button>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.5);
        backdrop-filter: blur(5px);
        -webkit-backdrop-filter: blur(5px);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10000;
        padding: 20px;
    }

    .modal-content {
        background: white;
        border-radius: 16px;
        width: 100%;
        max-width: 400px;
        box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    .modal-header {
        padding: 20px 20px 10px;
    }

    .modal-header h3 {
        margin: 0;
        font-size: 18px;
        color: #1f2937;
        font-weight: 700;
    }

    .modal-body {
        padding: 0 20px 20px;
        color: #4b5563;
        font-size: 15px;
        line-height: 1.5;
    }

    .modal-body p {
        margin: 0;
    }

    .modal-footer {
        display: flex;
        border-top: 1px solid #f3f4f6;
        background: #f9fafb;
        padding: 15px 20px;
        gap: 10px;
        justify-content: flex-end;
    }

    .modal-footer button {
        padding: 10px 16px;
        border-radius: 8px;
        font-size: 14px;
        font-weight: 600;
        cursor: pointer;
        border: none;
        transition: all 0.2s ease;
    }

    .btn-cancel {
        background: transparent;
        color: #4b5563;
    }

    .btn-cancel:hover {
        background: #f3f4f6;
    }

    .btn-confirm {
        background: #0ea5e9;
        color: white;
        box-shadow: 0 2px 4px rgba(14, 165, 233, 0.3);
    }

    .btn-confirm:hover {
        background: #0284c7;
    }

    /* Dark mode support */
    @media (prefers-color-scheme: dark) {
        .modal-content {
            background: #1f2937;
        }
        .modal-header h3 {
            color: #f9fafb;
        }
        .modal-body {
            color: #d1d5db;
        }
        .modal-footer {
            background: #111827;
            border-top-color: #374151;
        }
        .btn-cancel {
            color: #d1d5db;
        }
        .btn-cancel:hover {
            background: #374151;
        }
    }
</style>
