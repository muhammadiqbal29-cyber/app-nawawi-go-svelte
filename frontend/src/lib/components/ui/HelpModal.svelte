<script>
  import { fade, fly } from 'svelte/transition';
  import Icon from '$lib/components/ui/Icon.svelte';
  
  let { isOpen = $bindable(false), currentPath = '' } = $props();

  let activeAccordion = $state(0);

  function toggleAccordion(index) {
    if (activeAccordion === index) {
      activeAccordion = -1;
    } else {
      activeAccordion = index;
    }
  }

  // Konten Bantuan Kontekstual Berdasarkan Halaman
  let helpContent = $derived.by(() => {
    // 1. INPUT NILAI (List Kelas atau Detail Siswa)
    if (currentPath.includes('/guru/input-nilai') || currentPath.includes('/guru/input-diniyah')) {
      return {
        title: 'Bantuan: Input Nilai',
        icon: 'edit',
        items: [
          {
            title: 'Sistem Anti-Mati Lampu (Offline)',
            content: 'Anda dapat terus mengetik nilai meskipun internet mati. Data akan diamankan sementara di Brankas Lokal HP/Laptop Anda dan dikirim otomatis saat internet kembali.'
          },
          {
            title: 'Arti Ikon Awan',
            content: '<ul><li>🟢 <strong>Awan Hijau:</strong> Nilai aman di server.</li><li>🟡 <strong>Awan Kuning:</strong> Nilai sedang antre (internet lemah/putus). Biarkan saja.</li><li>⚪ <strong>Awan Abu:</strong> Belum ada nilai yang diubah.</li></ul>'
          },
          {
            title: 'Peringatan: Nilai Timpa-Menimpa',
            content: '<span style="color:#eab308;font-weight:600;">PENTING:</span> Pastikan hanya 1 GURU yang bertugas memasukkan nilai untuk kelas dan mapel yang sama. Jika 2 guru memasukkan nilai untuk anak yang sama saat offline, nilai yang terkirim belakangan akan menimpa nilai sebelumnya.'
          }
        ]
      };
    }
    
    // 2. RAPOR & KELULUSAN
    if (currentPath.includes('/guru/rapor')) {
      return {
        title: 'Bantuan: Pengelolaan Rapor',
        icon: 'fileText',
        items: [
          {
            title: 'Penghitungan Otomatis',
            content: 'Sistem akan secara otomatis mengakumulasi nilai dari input harian/semester. Anda hanya perlu memberikan Catatan Wali Kelas jika diperlukan.'
          },
          {
            title: 'Fitur Catatan Massal',
            content: 'Gunakan tombol "Catatan Massal" untuk memberikan catatan rapor dengan cepat ke banyak murid sekaligus tanpa membuka profil mereka satu-satu.'
          }
        ]
      };
    }

    // 3. DEFAULT (Dashboard dll)
    return {
      title: 'Panduan Umum Aplikasi',
      icon: 'info',
      items: [
        {
          title: 'Navigasi Jenjang & Tahun',
          content: 'Anda dapat berpindah Jenjang (SD/SMP/SMA) atau merubah Tahun Ajaran melalui menu dropdown di Bilah Atas (Header).'
        },
        {
          title: 'Background Auto-Sync (Sinkronisasi Awal)',
          content: 'Setiap Anda login atau berpindah jenjang saat sedang online, sistem secara cerdas akan mengunduh seluruh data siswa ke HP Anda agar siap digunakan secara offline kapan saja.'
        },
        {
          title: 'Peringatan: Hindari Tombol Refresh (F5)',
          content: '<span style="color:#eab308;font-weight:600;">TIPS KUOTA:</span> Jangan menekan tombol Refresh/F5 di browser jika tidak mendesak. Menekan Refresh akan memaksa sistem mengunduh ulang seluruh data dari awal (boros kuota internet). Cukup gunakan menu di sebelah kiri untuk berpindah halaman.'
        }
      ]
    };
  });

  function close() {
    isOpen = false;
  }
</script>

{#if isOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="modal-backdrop" transition:fade={{duration: 200}} onclick={close}>
    <div class="modal-content" transition:fly={{y: 50, duration: 300}} onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div class="modal-title">
          <div class="icon-wrap">
            <Icon name={helpContent.icon} size={20} />
          </div>
          <h2>{helpContent.title}</h2>
        </div>
        <button class="close-btn" onclick={close}>
          <Icon name="x" size={20} />
        </button>
      </div>

      <div class="modal-body">
        <div class="accordion-list">
          {#each helpContent.items as item, i}
            <div class="accordion-item" class:active={activeAccordion === i}>
              <button class="accordion-header" onclick={() => toggleAccordion(i)}>
                <h3>{item.title}</h3>
                <div class="chevron">
                  <Icon name="chevronDown" size={16} />
                </div>
              </button>
              {#if activeAccordion === i}
                <div class="accordion-body" transition:fade={{duration: 150}}>
                  <div class="accordion-inner">
                    {@html item.content}
                  </div>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(15, 23, 42, 0.6);
    backdrop-filter: blur(4px);
    z-index: 99999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }

  .modal-content {
    background: white;
    width: 100%;
    max-width: 500px;
    border-radius: 20px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    max-height: 85vh;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    background: #f8fafc;
    border-bottom: 1px solid #e2e8f0;
  }

  .modal-title {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .icon-wrap {
    width: 36px;
    height: 36px;
    background: #eff6ff;
    color: #3b82f6;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-title h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #0f172a;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #64748b;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #e2e8f0;
    color: #0f172a;
  }

  .modal-body {
    padding: 24px;
    overflow-y: auto;
  }

  .accordion-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .accordion-item {
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    overflow: hidden;
    background: white;
    transition: all 0.2s;
  }

  .accordion-item.active {
    border-color: #3b82f6;
    box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.1);
  }

  .accordion-header {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
  }

  .accordion-header h3 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: #1e293b;
  }

  .chevron {
    color: #64748b;
    transition: transform 0.3s ease;
    display: flex;
  }

  .accordion-item.active .chevron {
    transform: rotate(180deg);
    color: #3b82f6;
  }

  .accordion-body {
    background: #f8fafc;
    border-top: 1px solid #f1f5f9;
  }

  .accordion-inner {
    padding: 16px;
    font-size: 14px;
    line-height: 1.6;
    color: #475569;
  }

  /* Support for unordered lists inside content */
  :global(.accordion-inner ul) {
    margin: 8px 0 0 0;
    padding-left: 20px;
  }
  
  :global(.accordion-inner li) {
    margin-bottom: 6px;
  }
</style>
