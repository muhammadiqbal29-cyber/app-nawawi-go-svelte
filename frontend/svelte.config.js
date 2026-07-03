import adapter from '@sveltejs/adapter-node';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  kit: {
    // Menggunakan adapter-node untuk Native/PM2 Deployment
    adapter: adapter()
  }
};

export default config;
