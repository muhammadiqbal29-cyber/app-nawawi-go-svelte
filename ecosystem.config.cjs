module.exports = {
  apps: [
    {
      name: 'siakad-backend',
      script: './server',
      cwd: './backend',
      instances: 1,
      exec_mode: 'fork',
      env: {
        APP_ENV: 'production',
        APP_PORT: '8080',
        SSL_CERT_FILE: '/data/data/com.termux/files/usr/etc/tls/cert.pem'
      }
    },
    {
      name: 'siakad-frontend',
      script: 'node',
      args: 'build/index.js',
      cwd: './frontend',
      instances: 1,
      exec_mode: 'fork',
      env: {
        NODE_ENV: 'production',
        PORT: 3000
      }
    },
    {
      name: 'siakad-mobile',
      script: 'npx',
      args: '--yes serve -s build -l 4173',
      cwd: './mobile',
      instances: 1,
      exec_mode: 'fork',
      env: {
        NODE_ENV: 'production'
      }
    }
  ]
};
