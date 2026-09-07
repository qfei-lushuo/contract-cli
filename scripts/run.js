#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const { recoverWindowsBinary } = require("./windows-recovery.js");

const binaryName = process.platform === "win32" ? "contract-cli.exe" : "contract-cli";
const binaryPath = path.join(__dirname, "..", "bin", binaryName);

if (process.platform === "win32") {
  recoverWindowsBinary(binaryPath);
}

if (!fs.existsSync(binaryPath)) {
  console.error(
    `Error: contract-cli binary not found at ${binaryPath}\n\n` +
      "The npm postinstall step may have been skipped.\n" +
      `Run: node "${path.join(__dirname, "install.js")}"\n`
  );
  process.exit(1);
}

try {
  execFileSync(binaryPath, process.argv.slice(2), { stdio: "inherit" });
} catch (error) {
  process.exit(error.status || 1);
}
