const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/input-nilai/keseharian/kelas/[id]/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

// Replace submitIqob
const submitRegex = /async function submitIqob\(\) \{[\s\S]*?\} finally \{\n\s*iqobFormLoading = false;\n\s*\}\n\s*\}/;
const newSubmit = `function tambahPendingIqob() {
    if (!iqobFormJenisId || !iqobFormTanggal) {
      iqobFormError = 'Pilih jenis iqob dan tanggal terlebih dahulu';
      return;
    }
    iqobFormError = '';
    pendingIqobList = [...pendingIqobList, {
      jenis_iqob_id: iqobFormJenisId,
      nama: iqobSelectedNama,
      tanggal: iqobFormTanggal,
      poin: iqobFormPoin
    }];
    iqobFormJenisId = '';
    iqobFormPoin = 0;
    iqobSearchQuery = '';
    iqobShowDropdown = true;
    iqobSelectedNama = '';
  }

  function removePendingIqob(index) {
    pendingIqobList = pendingIqobList.filter((_, i) => i !== index);
  }

  async function submitIqob() {
    if (pendingIqobList.length === 0) {
      iqobFormError = 'Tambahkan setidaknya satu iqob ke daftar';
      return;
    }
    iqobFormLoading = true;
    iqobFormError = '';

    try {
      const url = new URL(\`\${apiUrl}/api/guru/iqob/catatan\`);
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);

      const promises = pendingIqobList.map(item => {
        return fetch(url.toString(), {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            siswa_id: Number(iqobFormSiswaId),
            sekolah_id: Number(sekolahId),
            jenis_iqob_id: Number(item.jenis_iqob_id),
            tanggal: item.tanggal,
            poin: Number(item.poin)
          })
        }).then(async r => {
          if (!r.ok) {
            const b = await r.json().catch(() => ({}));
            throw new Error(b.message || \`Gagal menambah iqob: \${item.nama}\`);
          }
        });
      });

      await Promise.all(promises);

      toastStore.show('Iqob berhasil dicatat', 'success');
      pendingIqobList = [];
      await loadIqobKelas();
    } catch (err) {
      iqobFormError = err.message;
    } finally {
      iqobFormLoading = false;
    }
  }`;

content = content.replace(submitRegex, newSubmit);

// UI Replacement for Iqob
const uiRegex = /\{:else\}\n\s*<input type="text" class="input" placeholder="Ketik jenis iqob\.\.\." bind:value=\{iqobSearchQuery\} \/>[\s\S]*?\{#if iqobFormError\}\n\s*<div class="alert-error mt-2">\{iqobFormError\}<\/div>\n\s*\{\/if\}\n\s*<\/div>/;

const newUI = `{:else}
                  {#if recommendedIqob}
                    <div style="margin-bottom: 0.5rem; background: #e0f2fe; border: 1px solid #bae6fd; padding: 0.5rem; border-radius: 4px; display: flex; align-items: center; justify-content: space-between;">
                      <div>
                        <span style="font-size: 0.8rem; font-weight: 600; color: #0284c7; display: block; margin-bottom: 2px;">💡 Rekomendasi Iqob (Poin Pelanggaran Tersisa: {getPoinAkhir(selectedStudent.siswa_id) - pendingIqobPoints}):</span>
                        <span style="font-size: 0.85rem; color: #0f172a; font-weight: 500;">{recommendedIqob.nama}</span>
                        <span style="color: #16a34a; font-size: 0.85rem; font-weight: 600; margin-left: 4px;">(-{recommendedIqob.poin})</span>
                      </div>
                      <button class="btn-secondary" style="padding: 0.25rem 0.5rem; font-size: 0.75rem;" onclick={applyRecommendedIqob}>
                        Gunakan
                      </button>
                    </div>
                  {/if}
                  <input type="text" class="input" placeholder="Ketik jenis iqob..." bind:value={iqobSearchQuery} />
                  {#if iqobSearchQuery.length > 0}
                    <div style="position: absolute; top: 100%; left: 0; right: 0; background: white; border: 1px solid #cbd5e1; border-radius: 4px; max-height: 250px; overflow-y: auto; z-index: 50; box-shadow: 0 4px 12px rgba(0,0,0,0.15); margin-top: 4px;">
                      {#if iqobFilteredOptions.length === 0}
                        <div style="padding: 0.75rem; color: #64748b; font-size: 0.9rem; text-align: center;">Tidak ada hasil ditemukan.</div>
                      {:else}
                        {#each iqobFilteredOptions as j}
                          <!-- svelte-ignore a11y_click_events_have_key_events -->
                          <!-- svelte-ignore a11y_no_static_element_interactions -->
                          <div
                            style="padding: 0.75rem; border-bottom: 1px solid #f1f5f9; cursor: pointer; font-size: 0.9rem; transition: background 0.1s;"
                            onmouseover={(e) => e.currentTarget.style.background = '#f1f5f9'}
                            onmouseout={(e) => e.currentTarget.style.background = 'transparent'}
                            onclick={() => selectIqob(j)}
                          >
                            <div style="font-weight: 500; color: #0f172a; margin-bottom: 0.25rem;">{j.nama}</div>
                            <div style="color: #16a34a; font-size: 0.8rem; font-weight: 700;">-{j.poin} Poin</div>
                          </div>
                        {/each}
                      {/if}
                    </div>
                  {/if}
                {/if}
              </div>
              <div class="form-group" style="flex: 1; min-width: 120px;">
                <label for="iqob-tgl-modal">Tanggal</label>
                <input id="iqob-tgl-modal" type="date" class="input" bind:value={iqobFormTanggal} />
              </div>
              <div class="form-group" style="flex: 1; min-width: 100px;">
                <label for="iqob-poin-modal">Poin (-)</label>
                <input id="iqob-poin-modal" type="number" step="0.1" class="input" bind:value={iqobFormPoin} />
              </div>
            </div>

            <div style="display: flex; justify-content: flex-end; margin-top: 1rem;">
              <button class="btn-secondary" onclick={tambahPendingIqob} disabled={iqobFormLoading}>
                <Icon name="plus" size={14} /> Tambah ke Daftar
              </button>
            </div>
            {#if iqobFormError}
              <div class="alert-error mt-2">{iqobFormError}</div>
            {/if}

            {#if pendingIqobList.length > 0}
              <div style="margin-top: 1.5rem; background: #ffffff; border: 1px dashed #cbd5e1; border-radius: 8px; padding: 1rem;">
                <h5 style="margin: 0 0 0.75rem 0; font-size: 0.9rem; color: #475569;">Daftar Iqob Baru (Belum Disimpan)</h5>
                {#each pendingIqobList as p, i}
                  <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; border-bottom: 1px solid #f1f5f9;">
                    <div>
                      <span style="font-weight: 500; font-size: 0.9rem;">{p.nama}</span>
                      <span style="color: #64748b; font-size: 0.8rem; margin-left: 8px;">{formatDate(p.tanggal)}</span>
                      <span style="color: #16a34a; font-size: 0.8rem; font-weight: 600; margin-left: 8px;">(-{p.poin})</span>
                    </div>
                    <button class="btn-icon-mini text-red" title="Hapus" onclick={() => removePendingIqob(i)}>
                      <Icon name="x" size={14} />
                    </button>
                  </div>
                {/each}
                <div style="display: flex; justify-content: flex-end; margin-top: 1rem;">
                  <button class="btn-primary" onclick={submitIqob} disabled={iqobFormLoading}>
                    {iqobFormLoading ? 'Menyimpan...' : \`Simpan Semua (\${pendingIqobList.length})\`}
                  </button>
                </div>
              </div>
            {/if}
          </div>`;

content = content.replace(uiRegex, newUI);

fs.writeFileSync(file, content);
console.log('Fixed');
