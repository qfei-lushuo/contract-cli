"use strict";

const { execFileSync } = require("child_process");
const fs = require("fs");

function restoreOldBinary(binaryPath, oldBinaryPath) {
  try {
    if (fs.existsSync(binaryPath)) {
      fs.rmSync(binaryPath, { force: true });
    }
    fs.renameSync(oldBinaryPath, binaryPath);
    return true;
  } catch (_) {
    return false;
  }
}

// Recover an interrupted Windows update, or remove the backup only after the
// replacement answers --version successfully. Kept platform-independent so
// the state transitions are testable on macOS/Linux CI too.
function recoverWindowsBinary(binaryPath, runBinary = execFileSync) {
  const oldBinaryPath = binaryPath + ".old";
  if (!fs.existsSync(oldBinaryPath)) {
    return "unchanged";
  }
  if (!fs.existsSync(binaryPath)) {
    return restoreOldBinary(binaryPath, oldBinaryPath) ? "restored" : "restore_failed";
  }
  try {
    runBinary(binaryPath, ["--version"], { stdio: "ignore", timeout: 10000 });
    try {
      fs.rmSync(oldBinaryPath, { force: true });
    } catch (_) {
      // Best-effort cleanup; the verified replacement can still run.
    }
    return "verified";
  } catch (_) {
    return restoreOldBinary(binaryPath, oldBinaryPath) ? "restored" : "restore_failed";
  }
}

module.exports = { recoverWindowsBinary };
