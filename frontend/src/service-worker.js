import { build, files, version } from '$service-worker';

// Create a unique cache name for this deployment
const CACHE = `siakad-cache-${version}`;

// We cache the build files (JS/CSS) and static files (manifest, icons, etc.)
const ASSETS = [
	...build,
	...files
];

self.addEventListener('install', (event) => {
	// Create a new cache and add all files to it
	async function addFilesToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	}

	event.waitUntil(addFilesToCache());
});

self.addEventListener('activate', (event) => {
	// Tell the active service worker to take control of the page immediately
	event.waitUntil(self.clients.claim());
	
	// Remove previous cached data from disk
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	event.waitUntil(deleteOldCaches());
});

self.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') return;

	const url = new URL(event.request.url);

	// Bypass ALL API endpoints from Service Worker caching
	// We rely on localStorage for offline fallback, not SW cache!
	if (url.pathname.startsWith('/api/')) {
		return;
	}

	if (url.origin === location.origin && url.pathname === '/login') {
		return;
	}
	async function respond() {
		const url = new URL(event.request.url);
		const cache = await caches.open(CACHE);

		// CACHE-FIRST untuk file statis dan build (CSS, JS, Fonts, Images)
		// Ini membuat aplikasi SANGAT CEPAT dan menghemat bandwidth
		if (ASSETS.includes(url.pathname)) {
			const cachedAsset = await cache.match(url.pathname);
			if (cachedAsset) {
				return cachedAsset;
			}
		}

		// NETWORK-FIRST untuk API dan Halaman HTML
		// Memastikan data yang didapat selalu terbaru dari server,
		// tapi mundur ke cache (offline) jika sinyal putus
		try {
			const response = await fetch(event.request);

			// if we're offline, fetch can return a value that is not a Response
			// instead of throwing - and we can't pass this non-Response to respondWith
			if (!(response instanceof Response)) {
				throw new Error('invalid response from fetch');
			}

			if (response.status === 200) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch (err) {
			const cachedResponse = await cache.match(event.request);
			if (cachedResponse) {
				return cachedResponse;
			}
            
            // If totally offline and not in cache, let it fail (or return a custom offline page)
			throw err;
		}
	}

	event.respondWith(respond());
});

self.addEventListener('message', (event) => {
	if (event.data && event.data.type === 'SKIP_WAITING') {
		self.skipWaiting();
	}
});
