<script>
  import InputNilai from '$lib/components/guru/InputNilai.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

  let hasAccess = $derived(
    guruState.activeSekolahId &&
      guruState.aksesMap[guruState.activeSekolahId] &&
      (guruState.aksesMap[guruState.activeSekolahId].includes('input_tahfiz') || guruState.aksesMap[guruState.activeSekolahId].includes('paket_alquran'))
  );
</script>

{#if !guruState.activeSekolahId}
  <div class="empty-state">
    <Icon name="grid" size={48} />
    <h3>Pilih Jenjang Sekolah</h3>
    <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas.</p>
  </div>
{:else if !hasAccess}
  <div class="empty-state">
    <Icon name="alertCircle" size={48} />
    <h3>Akses Ditolak</h3>
    <p>
      Anda tidak memiliki hak akses <b>Input Nilai</b> untuk jenjang {guruState.activeSekolahName}.
    </p>
  </div>
{:else}
  {#key guruState.activeSekolahId}
    <InputNilai sekolahId={guruState.activeSekolahId} />
  {/key}
{/if}

<style>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
    color: #64748b;
    background: white;
    border-radius: 1rem;
    border: 1px solid #e2e8f0;
  }
  .empty-state h3 {
    margin: 0;
    color: #0f172a;
  }
</style>
