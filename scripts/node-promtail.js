const fs = require('fs');
const http = require('http');
const { spawn } = require('child_process');
const path = require('path');

const LOKI_URL = 'http://localhost:3100/loki/api/v1/push';
const LOG_DIR = '/data/data/com.termux/files/home/.pm2/logs';

let counter = 0;

function pushToLoki(job, line) {
  counter++;
  const nanoTime = (Date.now() * 1000000) + counter;
  const payload = JSON.stringify({
    streams: [{
      stream: { job },
      values: [[ nanoTime.toString(), line ]]
    }]
  });

  const req = http.request(LOKI_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(payload)
    }
  }, (res) => {
     if (res.statusCode >= 400) {
        let err = '';
        res.on('data', d => err += d);
        res.on('end', () => console.log('Loki Error:', err));
     }
  });
  
  req.on('error', (e) => console.log('Req error', e)); 
  req.write(payload);
  req.end();
}

const files = fs.readdirSync(LOG_DIR);
files.forEach(file => {
  if (!file.endsWith('.log')) return;
  const filePath = path.join(LOG_DIR, file);
  
  let jobName = file.replace('.log', '').replace(/-out(-\d+)?$/, '').replace(/-error(-\d+)?$/, '');
  if (file.includes('error')) jobName += '-error';

  const tail = spawn('tail', ['-F', '-n', '50', filePath]);
  
  tail.stdout.on('data', (data) => {
    const lines = data.toString().split('\n');
    lines.forEach(line => {
      if (line.trim()) pushToLoki(jobName, line);
    });
  });
});
