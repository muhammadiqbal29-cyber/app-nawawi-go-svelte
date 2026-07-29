class SyncManager {
  queue = $state([]);
  isOffline = $state(false);
  isSyncing = $state(false);
  lastError = $state(null);
  isUnauthorized = $state(false);

  constructor() {
    if (typeof window !== 'undefined') {
      this.isOffline = !navigator.onLine;
      
      window.addEventListener('online', () => { 
        this.isOffline = false; 
        this.processQueue(); 
      });
      
      window.addEventListener('offline', () => { 
        this.isOffline = true; 
      });

      // 1. Advanced Stability: visibilitychange listener (wake up from sleep/background)
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible' && !this.isOffline && !this.isUnauthorized) {
          this.processQueue();
        }
      });

      // 2. Advanced Stability: beforeunload listener (warn if closing with pending queue)
      window.addEventListener('beforeunload', (e) => {
        if (this.queue.length > 0) {
          e.preventDefault();
          e.returnValue = 'Anda memiliki data yang belum tersimpan ke server. Yakin ingin keluar?';
        }
      });

      const stored = localStorage.getItem('siakad_sync_queue');
      if (stored) {
        try { 
          this.queue = JSON.parse(stored); 
        } catch (e) { 
          this.queue = []; 
        }
      }
      
      setInterval(() => {
        if (!this.isOffline && !this.isSyncing && !this.isUnauthorized && this.queue.length > 0) {
          this.processQueue();
        }
      }, 5000);
    }
  }

  saveQueue() {
    if (typeof window !== 'undefined') {
      // 3. Advanced Stability: QuotaExceededError handling
      try {
        localStorage.setItem('siakad_sync_queue', JSON.stringify(this.queue));
      } catch (err) {
        if (err.name === 'QuotaExceededError' || err.code === 22) {
          this.lastError = 'Memori HP Penuh! Harap bersihkan sedikit memori.';
          alert('Memori penyimpanan perangkat Anda penuh. Tidak dapat menyimpan data cadangan offline. Harap hapus beberapa file/foto agar nilai tidak hilang.');
        } else {
          this.lastError = 'Gagal menyimpan antrean lokal.';
        }
      }
    }
  }

  enqueue(url, payload, options = {}) {
    const item = {
      id: Date.now() + Math.random().toString(36).substr(2, 9),
      url,
      payload,
      options,
      timestamp: Date.now()
    };
    
    // Deduplication logic:
    // Prevent overwriting distinct records of the same student
    const siswaId = payload.siswa_id;
    let replaced = false;
    
    if (siswaId) {
      for (let i = 0; i < this.queue.length; i++) {
        if (this.queue[i].url === url && this.queue[i].payload?.siswa_id === siswaId) {
          // Distinct mapel diniyah
          if (payload.mapel_diniyah_id && this.queue[i].payload?.mapel_diniyah_id !== payload.mapel_diniyah_id) {
            continue;
          }
          // Distinct surah_id (Tahfiz V2)
          if (payload.surah_id && this.queue[i].payload?.surah_id !== payload.surah_id) {
            continue;
          }
          // Distinct nama_materi (Tahfiz V1, V2, Tasmi V1, V2)
          if (payload.nama_materi && this.queue[i].payload?.nama_materi !== payload.nama_materi) {
            continue;
          }
          // Distinct jenis_pelanggaran_id
          if (payload.jenis_pelanggaran_id && this.queue[i].payload?.jenis_pelanggaran_id !== payload.jenis_pelanggaran_id) {
            continue;
          }
          // Distinct jenis_iqob_id
          if (payload.jenis_iqob_id && this.queue[i].payload?.jenis_iqob_id !== payload.jenis_iqob_id) {
            continue;
          }
          
          this.queue[i] = item;
          replaced = true;
          break;
        }
      }
    }
    
    if (!replaced) {
      this.queue.push(item);
    }
    
    this.saveQueue();
    this.processQueue();
  }

  async processQueue() {
    if (this.isSyncing || this.queue.length === 0 || this.isOffline || this.isUnauthorized) return;
    
    this.isSyncing = true;
    this.lastError = null;

    while (this.queue.length > 0) {
      if (this.isOffline || this.isUnauthorized) break;
      const item = this.queue[0];
      
      try {
        const fetchOptions = {
          method: item.options.method || 'POST',
          headers: { 'Content-Type': 'application/json', ...item.options.headers },
          body: JSON.stringify(item.payload),
          credentials: 'include'
        };

        const res = await window.fetch(item.url, fetchOptions);
        
        if (res.status === 401 || res.status === 403) {
           // Token expired or unauthorized! DO NOT delete the queue!
           this.isUnauthorized = true;
           this.lastError = "Sesi berakhir. Harap login kembali.";
           break;
        }

        if (res.status === 429) {
           // Terlalu banyak request, sistem rate limiter menolak. Break loop agar tidak kehapus
           this.lastError = "Terlalu banyak request. Menunggu sejenak...";
           break;
        }

        // If success or client error (except 401/403/429) 
        if (res.ok || (res.status >= 400 && res.status < 500)) {
           this.queue.shift();
           this.saveQueue();
        } else {
           // Server error (5xx), keep in queue to retry later
           this.lastError = `Server error ${res.status}`;
           break;
        }
      } catch (err) {
        // Network error
        this.isOffline = true;
        this.lastError = "Jaringan terputus. Menunggu koneksi...";
        break;
      }
      
      // Memberikan jeda (delay) antar request untuk menghindari spamming / 429 dari Cloudflare
      await new Promise(r => setTimeout(r, 300));
    }
    
    this.isSyncing = false;
  }
}

export const syncManager = new SyncManager();
