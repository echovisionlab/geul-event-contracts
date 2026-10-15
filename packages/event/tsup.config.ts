import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm"],
  clean: true,
  external: ["@bufbuild/protobuf"],
  // Bundle internal files but keep external dependencies external
  noExternal: [],
});
