const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const path = require("node:path");

for (const directory of process.argv.slice(2)) {
  const { name, version } = require(path.resolve(directory, "package.json"));
  const published = JSON.parse(
    execFileSync("npm", ["view", `${name}@${version}`, "--json"], {
      encoding: "utf8",
    }),
  );
  assert.equal(published.name, name);
  assert.equal(published.version, version);
  assert.match(published.dist.integrity, /^sha512-[A-Za-z0-9+/]+={0,2}$/);
  assert.equal(new URL(published.dist.tarball).hostname, "npm.pkg.github.com");
  console.log(JSON.stringify({ name, version, dist: published.dist }, null, 2));
}
