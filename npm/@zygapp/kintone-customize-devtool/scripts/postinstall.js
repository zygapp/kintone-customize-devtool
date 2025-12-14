const fs = require('fs');
const path = require('path');

const PLATFORMS = {
  'darwin-x64': 'darwin-x64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-x64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'win32-x64',
  'win32-arm64': 'win32-arm64',
};

function getPlatformDir() {
  const platform = process.platform;
  const arch = process.arch === 'x64' ? 'x64' : 'arm64';
  const key = `${platform}-${arch}`;
  return PLATFORMS[key];
}

function main() {
  const platformDir = getPlatformDir();

  if (!platformDir) {
    console.error(`kcdev: unsupported platform: ${process.platform}-${process.arch}`);
    process.exit(1);
  }

  const ext = process.platform === 'win32' ? '.exe' : '';
  const binaryPath = path.join(__dirname, '..', 'bin', platformDir, `kcdev${ext}`);

  if (!fs.existsSync(binaryPath)) {
    console.error(`kcdev: binary not found: ${binaryPath}`);
    process.exit(1);
  }

  // バイナリパスを記録
  const binDir = path.join(__dirname, '..', 'bin');
  const pathFile = path.join(binDir, '.binary-path');
  fs.writeFileSync(pathFile, binaryPath);

  console.log(`kcdev: installed for ${process.platform}-${process.arch}`);
}

main();
