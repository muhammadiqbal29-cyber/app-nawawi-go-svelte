import sharp from 'sharp';
import fs from 'fs';

async function generateIcons() {
  const source = '../logo.png';
  if (!fs.existsSync(source)) {
    console.error("Source image not found!");
    return;
  }
  
  console.log("Resizing icon-192.png...");
  await sharp(source).resize(192, 192).toFile('./static/icon-192.png');
  
  console.log("Resizing icon-512.png...");
  await sharp(source).resize(512, 512).toFile('./static/icon-512.png');
  
  console.log("Resizing icon-512-maskable.png...");
  await sharp(source)
    .resize(512, 512, { fit: 'contain', background: { r: 255, g: 255, b: 255, alpha: 1 } })
    .toFile('./static/icon-512-maskable.png');
    
  console.log("Resizing favicon.png...");
  await sharp(source).resize(64, 64).toFile('./static/favicon.png');
  
  console.log("Icons generated successfully.");
}

generateIcons().catch(console.error);
