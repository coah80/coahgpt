const https = require("https");
const fs = require("fs");
const path = require("path");

const BASE_URL = "https://coahgpt.com/releases";

const platform = process.platform === "darwin" ? "darwin" : "linux";
const arch = process.arch === "arm64" ? "arm64" : "amd64";
const binName = `coahgpt_${platform}_${arch}`;
const url = `${BASE_URL}/${binName}`;

const vendorDir = path.join(__dirname, "..", "vendor");
const binPath = path.join(vendorDir, binName);

fs.mkdirSync(vendorDir, { recursive: true });

console.log(`\n  /\\_/\\`);
console.log(`  ( o.o )  downloading coahGPT...`);
console.log(`  > ^ <\n`);

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          download(res.headers.location, dest).then(resolve).catch(reject);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: HTTP ${res.statusCode}`));
          return;
        }
        res.pipe(file);
        file.on("finish", () => {
          file.close();
          fs.chmodSync(dest, 0o755);
          resolve();
        });
      })
      .on("error", reject);
  });
}

download(url, binPath)
  .then(() => {
    console.log("  coahGPT installed successfully!");
    console.log(`\n  /\\_/\\`);
    console.log(`  ( ^.^ )  meow. run 'coahgpt' to start.`);
    console.log(`  > ^ <\n`);
  })
  .catch((err) => {
    console.error(`  /\\_/\\`);
    console.error(`  ( x.x )  install failed: ${err.message}`);
    console.error(`  > ^ <\n`);
    process.exit(1);
  });
