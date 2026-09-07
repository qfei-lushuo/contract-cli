"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const { recoverWindowsBinary } = require("./windows-recovery.js");

function withBinaryFiles(callback) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "contract-cli-run-test-"));
  const binary = path.join(dir, "contract-cli.exe");
  try {
    callback(binary);
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

test("restores .old when the replacement is missing", () => {
  withBinaryFiles((binary) => {
    fs.writeFileSync(binary + ".old", "previous");
    assert.equal(recoverWindowsBinary(binary), "restored");
    assert.equal(fs.readFileSync(binary, "utf8"), "previous");
    assert.equal(fs.existsSync(binary + ".old"), false);
  });
});

test("keeps a healthy replacement and removes .old", () => {
  withBinaryFiles((binary) => {
    fs.writeFileSync(binary, "new");
    fs.writeFileSync(binary + ".old", "previous");
    const run = (actual, args, options) => {
      assert.equal(actual, binary);
      assert.deepEqual(args, ["--version"]);
      assert.equal(options.timeout, 10000);
    };
    assert.equal(recoverWindowsBinary(binary, run), "verified");
    assert.equal(fs.readFileSync(binary, "utf8"), "new");
    assert.equal(fs.existsSync(binary + ".old"), false);
  });
});

test("restores .old when the replacement fails verification", () => {
  withBinaryFiles((binary) => {
    fs.writeFileSync(binary, "broken");
    fs.writeFileSync(binary + ".old", "previous");
    const run = () => {
      throw new Error("cannot execute");
    };
    assert.equal(recoverWindowsBinary(binary, run), "restored");
    assert.equal(fs.readFileSync(binary, "utf8"), "previous");
    assert.equal(fs.existsSync(binary + ".old"), false);
  });
});
