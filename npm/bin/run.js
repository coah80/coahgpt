#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const binDir = path.join(__dirname, "..", "vendor");
const platform = process.platform === "darwin" ? "darwin" : "linux";
const arch = process.arch === "arm64" ? "arm64" : "amd64";
const binName = `coahgpt_${platform}_${arch}`;
const binPath = path.join(binDir, binName);

if (!fs.existsSync(binPath)) {
  console.error("coahGPT binary not found. Run: npm rebuild coahgpt");
  process.exit(1);
}

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  process.exit(e.status || 1);
}
