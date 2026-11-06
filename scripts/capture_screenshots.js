const puppeteer = require('B:/workgit/node_modules/puppeteer');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const OUTPUT_DIR = path.resolve(__dirname, '../docs/images');
const FRAMES_DIR = path.resolve(__dirname, '../docs/images/frames');
const BASE_URL = 'http://127.0.0.1:8080';

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function run() {
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }
  if (!fs.existsSync(FRAMES_DIR)) {
    fs.mkdirSync(FRAMES_DIR, { recursive: true });
  }

  console.log('Launching headless browser...');
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage'],
  });

  // 1. Desktop Overview (1440x900)
  console.log('1. Capturing Desktop Overview...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(2000);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'restorecraft_overview.png') });
    console.log('✓ restorecraft_overview.png captured');
    await page.close();
  }

  // 2. Drill Detail Modal
  console.log('2. Capturing Drill Detail Modal...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(1500);

    const buttons = await page.$$('button');
    for (const btn of buttons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Details')) {
        await btn.click();
        break;
      }
    }
    await sleep(1000);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'restorecraft_drill_detail.png') });
    console.log('✓ restorecraft_drill_detail.png captured');
    await page.close();
  }

  // 3. Compliance Pack Modal
  console.log('3. Capturing Compliance Pack Modal...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(1500);

    const buttons = await page.$$('button');
    for (const btn of buttons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Details')) {
        await btn.click();
        break;
      }
    }
    await sleep(800);

    const compButtons = await page.$$('button');
    for (const btn of compButtons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Compliance Pack')) {
        await btn.click();
        break;
      }
    }
    await sleep(1200);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'restorecraft_compliance.png') });
    console.log('✓ restorecraft_compliance.png captured');
    await page.close();
  }

  // 4. Active Sandboxes & Interactive SQL Console
  console.log('4. Capturing Active Sandboxes & SQL Console...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(1500);

    // Open Sandboxes modal
    const sbButtons = await page.$$('button');
    for (const btn of sbButtons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Sandboxes')) {
        await btn.click();
        break;
      }
    }
    await sleep(1000);

    // Click Interactive SQL button
    const sqlButtons = await page.$$('button');
    for (const btn of sqlButtons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Interactive SQL')) {
        await btn.click();
        break;
      }
    }
    await sleep(800);

    // Click Run Query
    const runButtons = await page.$$('button');
    for (const btn of runButtons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Run Query')) {
        await btn.click();
        break;
      }
    }
    await sleep(1200);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'restorecraft_sql_console.png') });
    console.log('✓ restorecraft_sql_console.png captured');
    await page.close();
  }

  // 5. Time-Travel Bisect Modal
  console.log('5. Capturing Time-Travel Bisect Modal...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(1500);

    const bisectButtons = await page.$$('button');
    for (const btn of bisectButtons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Bisect')) {
        await btn.click();
        break;
      }
    }
    await sleep(800);

    const launchButtons = await page.$$('button');
    for (const btn of launchButtons) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Launch Chronological Bisect')) {
        await btn.click();
        break;
      }
    }
    await sleep(2500);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'restorecraft_bisect.png') });
    console.log('✓ restorecraft_bisect.png captured');
    await page.close();
  }

  // 6. Mobile Responsiveness (390x844)
  console.log('6. Capturing Mobile Responsiveness (iPhone 14/15/16 390x844)...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2, isMobile: true, hasTouch: true });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(1500);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'restorecraft_mobile.png') });
    console.log('✓ restorecraft_mobile.png captured');
    await page.close();
  }

  // 7. Record Animated Demo GIF frames
  console.log('7. Recording Demo GIF frames...');
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 800, deviceScaleFactor: 1 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await sleep(1500);

    let frameCount = 0;
    async function snap(times = 1, delay = 200) {
      for (let i = 0; i < times; i++) {
        const framePath = path.join(FRAMES_DIR, `frame_${String(frameCount++).padStart(4, '0')}.png`);
        await page.screenshot({ path: framePath });
        await sleep(delay);
      }
    }

    // Scene 1: Dashboard overview
    await snap(6, 250);

    // Scene 2: Open "Execute Drill" dialog
    const execBtns = await page.$$('button');
    for (const btn of execBtns) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Execute Drill')) {
        await btn.click();
        break;
      }
    }
    await snap(6, 250);

    // Scene 3: Launch drill
    const launchBtns = await page.$$('button');
    for (const btn of launchBtns) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Execute Recovery Drill')) {
        await btn.click();
        break;
      }
    }
    await snap(10, 200);

    // Scene 4: Open first drill details
    const rows = await page.$$('button');
    for (const btn of rows) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Details')) {
        await btn.click();
        break;
      }
    }
    await snap(8, 250);

    // Scene 5: Open Compliance Pack
    const compBtns = await page.$$('button');
    for (const btn of compBtns) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Compliance Pack')) {
        await btn.click();
        break;
      }
    }
    await snap(10, 250);

    // Scene 6: Close and open Sandboxes
    await page.keyboard.press('Escape');
    await sleep(400);

    const sbBtns = await page.$$('button');
    for (const btn of sbBtns) {
      const text = await page.evaluate((el) => el.textContent, btn);
      if (text && text.includes('Sandboxes')) {
        await btn.click();
        break;
      }
    }
    await snap(8, 250);

    await page.close();
    console.log(`✓ Captured ${frameCount} frames for demo GIF.`);
  }

  await browser.close();

  // 8. Compile animated GIF using FFmpeg
  console.log('8. Encoding animated GIF with FFmpeg...');
  const palettePath = path.join(OUTPUT_DIR, 'palette.png');
  const gifPath = path.join(OUTPUT_DIR, 'restorecraft_demo.gif');

  const ffmpegCmd1 = `ffmpeg -y -framerate 4 -i "${FRAMES_DIR}/frame_%04d.png" -vf "fps=4,scale=1000:-1:flags=lanczos,palettegen" "${palettePath}"`;
  execSync(ffmpegCmd1, { stdio: 'inherit' });

  const ffmpegCmd2 = `ffmpeg -y -framerate 4 -i "${FRAMES_DIR}/frame_%04d.png" -i "${palettePath}" -lavfi "fps=4,scale=1000:-1:flags=lanczos [x]; [x][1:v] paletteuse" "${gifPath}"`;
  execSync(ffmpegCmd2, { stdio: 'inherit' });

  // Cleanup frames and temporary palette
  fs.rmSync(FRAMES_DIR, { recursive: true, force: true });
  if (fs.existsSync(palettePath)) fs.unlinkSync(palettePath);

  console.log(`✓ Successfully compiled animated demo GIF: ${gifPath}`);
}

run().catch((err) => {
  console.error('Screenshot capture failed:', err);
  process.exit(1);
});
